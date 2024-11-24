package main

import (
	"context"
	"log"

	pb "github.com/salvatoreolivieri/commons/api"
	"google.golang.org/grpc"
)

type grpcHandler struct {
	pb.UnimplementedOrderServiceServer
	service OrdersService
}

func NewGRPCHandler(grpcServer *grpc.Server, service OrdersService) {
	handler := &grpcHandler{service: service}
	pb.RegisterOrderServiceServer(grpcServer, handler)
}

func (h *grpcHandler) CreateOrder(ctx context.Context, payload *pb.CreateOrderRequest) (*pb.Order, error) {

	err := h.service.ValidateOrder(ctx, payload)
	if err != nil {
		return nil, err
	}

	log.Printf("New order received! Order: %v", payload)

	order := &pb.Order{
		ID: payload.CustomerID,
	}

	return order, nil
}
