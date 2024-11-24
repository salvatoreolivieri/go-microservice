package main

import (
	"context"

	pb "github.com/salvatoreolivieri/commons/api"
)

type OrdersService interface {
	CreateOrder(context.Context) error
	ValidateOrder(context.Context, *pb.CreateOrderRequest) error
}

type OrdersStore interface {
	Create(context.Context) error
}
