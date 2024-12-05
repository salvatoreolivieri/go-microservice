package main

import (
	"context"

	pb "github.com/salvatoreolivieri/commons/api"
	"github.com/salvatoreolivieri/omsv-payments/gateway"
	"github.com/salvatoreolivieri/omsv-payments/processor"
)

type service struct {
	processor processor.PaymentProcessor
	gateway   gateway.OrdersGateway
}

func NewService(processor processor.PaymentProcessor, gateway gateway.OrdersGateway) *service {
	return &service{
		processor,
		gateway,
	}
}

func (s *service) CreatePayment(ctx context.Context, order *pb.Order) (string, error) {
	link, err := s.processor.CreatePaymentLink(order)
	if err != nil {
		return "", err
	}

	// update order with the link
	err = s.gateway.UpdateOrderAfterPaymentLink(ctx, order.ID, link)
	if err != nil {
		return "", err
	}

	return link, nil
}
