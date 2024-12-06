package main

import (
	"context"

	common "github.com/salvatoreolivieri/commons"
	pb "github.com/salvatoreolivieri/commons/api"
	"github.com/salvatoreolivieri/omsv-orders/gateway"
)

type service struct {
	store   OrdersStore
	gateway gateway.StockGateway
}

func NewService(store OrdersStore, gateway gateway.StockGateway) *service {
	return &service{store, gateway}
}

func (s *service) UpdateOrder(ctx context.Context, order *pb.Order) (*pb.Order, error) {
	err := s.store.Update(ctx, order.ID, order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *service) GetOrder(ctx context.Context, payload *pb.GetOrderRequest) (*pb.Order, error) {
	order, err := s.store.Get(ctx, payload.OrderID, payload.CustomerID)
	if err != nil {
		return nil, err
	}

	return order.ToProto(), nil
}

func (s *service) CreateOrder(ctx context.Context, payload *pb.CreateOrderRequest, items []*pb.Item) (*pb.Order, error) {
	orderID, err := s.store.Create(ctx, Order{
		CustomerID:  payload.CustomerID,
		Status:      "pending",
		Items:       items,
		PaymentLink: "",
	})
	if err != nil {
		return nil, err
	}

	order := &pb.Order{
		ID:         orderID.Hex(),
		CustomerID: payload.CustomerID,
		Status:     "pending",
		Items:      items,
	}

	return order, nil
}

func (s *service) ValidateOrder(ctx context.Context, order *pb.CreateOrderRequest) ([]*pb.Item, error) {
	if len(order.Items) == 0 {
		return nil, common.ErrNoItems
	}

	mergedItems := mergeItemsQuantities(order.Items)

	inStock, items, err := s.gateway.CheckIfItemIsInStock(ctx, order.CustomerID, mergedItems)
	if err != nil {
		return nil, err
	}
	if !inStock {
		return items, common.ErrNoStock
	}

	return items, nil
}

// mergeItemsQuantities merges a slice of ItemsWithQuantity objects by summing their quantities
// for items with the same ID. It ensures that no duplicate IDs exist in the resulting slice.
func mergeItemsQuantities(items []*pb.ItemsWithQuantity) []*pb.ItemsWithQuantity {

	// Create a slice to store the merged items.
	merged := make([]*pb.ItemsWithQuantity, 0)

	// Iterate over each item in the input slice.
	for _, item := range items {
		// Flag to check if the current item has already been merged.
		found := false

		// Iterate over the merged slice to check if the current item's ID already exists.
		for _, finalItem := range merged {
			// If the ID matches, update the quantity and mark it as found.
			if finalItem.ID == item.ID {
				finalItem.Quantity += item.Quantity
				found = true
				break // Exit the loop as the item has been merged.
			}
		}

		// If the item was not found in the merged slice, append it as a new entry.
		if !found {
			merged = append(merged, item)
		}
	}

	// Return the merged slice.
	return merged
}
