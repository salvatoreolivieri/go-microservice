package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/salvatoreolivieri/commons/api"
	"github.com/salvatoreolivieri/commons/broker"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

type grpcHandler struct {
	pb.UnimplementedOrderServiceServer
	service OrdersService
	channel *amqp.Channel
}

func NewGRPCHandler(grpcServer *grpc.Server, service OrdersService, channel *amqp.Channel) {
	handler := &grpcHandler{
		service: service,
		channel: channel,
	}

	pb.RegisterOrderServiceServer(grpcServer, handler)
}

func (h *grpcHandler) UpdateOrder(ctx context.Context, payload *pb.Order) (*pb.Order, error) {
	return h.service.UpdateOrder(ctx, payload)
}

func (h *grpcHandler) GetOrder(ctx context.Context, payload *pb.GetOrderRequest) (*pb.Order, error) {
	return h.service.GetOrder(ctx, payload)
}

func (h *grpcHandler) CreateOrder(ctx context.Context, payload *pb.CreateOrderRequest) (*pb.Order, error) {
	que, err := h.channel.QueueDeclare(broker.OrderCreatedEvent, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	tr := otel.Tracer("amqp")
	amqpContext, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - publish - %s", que.Name))
	defer messageSpan.End()

	_, err = h.service.ValidateOrder(amqpContext, payload)
	if err != nil {
		return nil, err
	}

	log.Printf("New order received! Order: %v", payload)

	items, err := h.service.ValidateOrder(amqpContext, payload)
	if err != nil {
		return nil, err
	}

	order, err := h.service.CreateOrder(amqpContext, payload, items)
	if err != nil {
		return nil, err
	}

	marshalledOrder, err := json.Marshal(order)
	if err != nil {
		log.Fatal(err)
	}

	// inject the header
	headers := broker.InjectAMQPHeaders(amqpContext)

	h.channel.PublishWithContext(amqpContext, "", que.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         marshalledOrder,
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
	})

	return order, nil
}
