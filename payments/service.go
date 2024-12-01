package main

import (
	"context"

	pb "github.com/salvatoreolivieri/commons/api"
	"github.com/salvatoreolivieri/omsv-payments/processor"
)

type service struct {
	processor processor.PaymentProcessor
}

func NewService(processor processor.PaymentProcessor) *service {
	return &service{
		processor,
	}
}

func (s *service) CreatePayment(ctx context.Context, order *pb.Order) (string, error) {
	link, err := s.processor.CreatePaymentLink(order)
	if err != nil {
		return "", err
	}

	return link, nil
}
