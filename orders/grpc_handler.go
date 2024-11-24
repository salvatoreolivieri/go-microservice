package main

import (
	"context"
	"log"

	pb "github.com/salvatoreolivieri/commons/api"
	"google.golang.org/grpc"
)

type grpcHandler struct {
	pb.UnimplementedOrderServiceServer
}

func NewGRPCHandler(grpcServer *grpc.Server) {
	handler := &grpcHandler{}
	pb.RegisterOrderServiceServer(grpcServer, handler)
}

func (h *grpcHandler) CreateOrder(ctx context.Context, payload *pb.CreateOrderRequest) (*pb.Order, error) {
	log.Printf("New order received! Order: %v", payload)

	order := &pb.Order{
		ID: payload.CustomerID,
	}
	log.Printf("local order: %v", order)

	return order, nil
}
