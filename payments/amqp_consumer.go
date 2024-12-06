package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/salvatoreolivieri/commons/api"
	"go.opentelemetry.io/otel"

	"github.com/salvatoreolivieri/commons/broker"
)

type consumer struct {
	service PaymentsService
}

func NewConsumer(service PaymentsService) *consumer {
	return &consumer{service}
}

func (c *consumer) Listen(channel *amqp.Channel) {

	que, err := channel.QueueDeclare(broker.OrderCreatedEvent, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	messages, err := channel.Consume(que.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	var forever chan struct{}

	go func() {
		for delivery := range messages {
			log.Printf("Received message: %v", delivery.Body)

			// extract the header
			ctx := broker.ExtractAMQPHeaders(context.Background(), delivery.Headers)

			tr := otel.Tracer("amqp")
			_, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - consume - %s", que.Name))

			order := &pb.Order{}
			if err := json.Unmarshal(delivery.Body, order); err != nil {
				delivery.Nack(false, false) // not acknowledging if the unmarshal fails
				log.Printf("ailed to unmarshal order: %v", err)
				continue
			}

			paymentLink, err := c.service.CreatePayment(context.Background(), order)
			if err != nil {
				log.Printf("failed to create payment: %v", err)

				if err := broker.HandleRetry(channel, &delivery); err != nil {
					log.Printf("error handling retry: %v", err)
					return
				}

				delivery.Nack(false, false)

				continue
			}

			messageSpan.AddEvent("payment.created")
			messageSpan.End()

			log.Printf("payment link created %s", paymentLink)
			delivery.Ack(false)
		}
	}()

	<-forever
}
