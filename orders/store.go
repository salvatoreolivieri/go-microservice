package main

import (
	"context"
	"errors"

	pb "github.com/salvatoreolivieri/commons/api"
)

var orders = make([]*pb.Order, 0)

type store struct {
	// add here our mongoDB
}

func NewStore() *store {
	return &store{}
}

func (s *store) Create(ctx context.Context, payload *pb.CreateOrderRequest, items []*pb.Item) (string, error) {

	id := "42"
	orders = append(orders, &pb.Order{
		ID:         id,
		CustomerID: payload.CustomerID,
		Status:     "pending",
		Items:      items,
	})

	return id, nil
}

func (s *store) Get(ctx context.Context, orderID, customerID string) (*pb.Order, error) {

	for _, order := range orders {
		if order.ID == orderID && order.CustomerID == customerID {
			return order, nil
		}
	}

	return nil, errors.New("order not found")
}
