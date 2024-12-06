package gateway

import (
	"context"

	pb "github.com/salvatoreolivieri/commons/api"
)

type KitchenGateway interface {
	UpdateOrder(context.Context, *pb.Order) error
}
