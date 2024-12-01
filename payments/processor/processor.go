package processor

import (
	pb "github.com/salvatoreolivieri/commons/api"
)

type PaymentProcessor interface {
	CreatePaymentLink(*pb.Order) (string, error)
}
