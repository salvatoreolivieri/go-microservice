package main

import (
	"context"

	pb "github.com/salvatoreolivieri/commons/api"
)

type service struct {
}

func NewService() *service {
	return &service{}
}

func (s *service) CreatePayment(ctx context.Context, order *pb.Order) (string, error) {
	// Implement your payment processing logic here
	// For example, you can use a payment gateway service to create a payment

	return "payment-link", nil
}
