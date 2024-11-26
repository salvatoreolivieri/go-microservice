package gateway

import (
	"context"

	pb "github.com/salvatoreolivieri/commons/api"
)

type OrdersGateway interface {
	CreateOrder(context.Context, *pb.CreateOrderRequest) (*pb.Order, error)
}
