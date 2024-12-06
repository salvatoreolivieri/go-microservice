package main

import (
	"context"
	"fmt"

	pb "github.com/salvatoreolivieri/commons/api"
	"go.opentelemetry.io/otel/trace"
)

type TelemetryMiddleware struct {
	next OrdersService
}

func NewTelemetryMiddleware(next OrdersService) OrdersService {
	return &TelemetryMiddleware{next}
}

func (s *TelemetryMiddleware) UpdateOrder(ctx context.Context, order *pb.Order) (*pb.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("UpdateOrder: %v", order))

	return s.next.UpdateOrder(ctx, order)
}

func (s *TelemetryMiddleware) GetOrder(ctx context.Context, payload *pb.GetOrderRequest) (*pb.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("GetOrder: %v", payload))

	return s.next.GetOrder(ctx, payload)
}

func (s *TelemetryMiddleware) CreateOrder(ctx context.Context, payload *pb.CreateOrderRequest, items []*pb.Item) (*pb.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("CreateOrder: %v", payload))

	return s.next.CreateOrder(ctx, payload, items)
}

func (s *TelemetryMiddleware) ValidateOrder(ctx context.Context, order *pb.CreateOrderRequest) ([]*pb.Item, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("ValidateOrder: %v", order))

	return s.next.ValidateOrder(ctx, order)
}
