package main

import (
	"context"

	pb "github.com/salvatoreolivieri/commons/api"
)

type PaymentsService interface {
	CreatePayment(context.Context, *pb.Order) (string, error)
}
