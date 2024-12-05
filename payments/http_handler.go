package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"fmt"
	"io"
	"net/http"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/salvatoreolivieri/commons/broker"

	pb "github.com/salvatoreolivieri/commons/api"
	// "github.com/salvatoreolivieri/commons/broker"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type PaymentHTTPHandler struct {
	channel *amqp.Channel
}

func NewPaymentHTTPHandler(channel *amqp.Channel) *PaymentHTTPHandler {
	return &PaymentHTTPHandler{channel}
}

func (h *PaymentHTTPHandler) registerRoutes(router *http.ServeMux) {
	router.HandleFunc("/webhook", h.handleCheckoutWebhook)
}

func (h *PaymentHTTPHandler) handleCheckoutWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	fmt.Fprintf(os.Stdout, "Got body: %s\n", payload)

	// Pass the request body and Stripe-Signature header to ConstructEvent, along with the webhook signing key.
	event, err := webhook.ConstructEventWithOptions(payload, r.Header.Get("Stripe-Signature"), endpointStripSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if session.PaymentStatus == "paid" {
			log.Printf("Payment for Checkout Session %v succeeded", session.ID)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			orderID := session.Metadata["orderID"]
			customerID := session.Metadata["customerID"]

			order := &pb.Order{
				ID:          orderID,
				CustomerID:  customerID,
				Status:      "paid",
				PaymentLink: "",
			}

			marshalledOrder, err := json.Marshal(order)
			if err != nil {
				// TODO: add error
			}

			// Fanout messages to other services
			h.channel.PublishWithContext(ctx, broker.OrderPaidEvent, "", false, false, amqp.Publishing{
				ContentType:  "application/json",
				Body:         marshalledOrder,
				DeliveryMode: amqp.Persistent,
			})
		}
	}

	w.WriteHeader(http.StatusOK)

}
