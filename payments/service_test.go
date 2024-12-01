package main

import (
	"context"
	"testing"

	"github.com/salvatoreolivieri/commons/api"
	"github.com/salvatoreolivieri/omsv-payments/processor/inmem"
)

func TestService(t *testing.T) {
	processor := inmem.NewInmem()
	service := NewService(processor)

	t.Run("Should create a payment link", func(t *testing.T) {
		link, err := service.CreatePayment(context.Background(), &api.Order{})
		if err != nil {
			t.Errorf("CreatePayment() error = %v, want nil", err)
		}

		if link == "" {
			t.Error("CreatePayment() link is empty")
		}
	})
}
