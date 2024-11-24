package main

import (
	"context"
	"log"

	common "github.com/salvatoreolivieri/commons"
	pb "github.com/salvatoreolivieri/commons/api"
)

type service struct {
	store OrdersStore
}

func NewService(store OrdersStore) *service {

	return &service{store}
}

func (s *service) CreateOrder(context.Context) error {
	return nil
}

func (s *service) ValidateOrder(ctx context.Context, order *pb.CreateOrderRequest) error {
	if len(order.Items) == 0 {
		return common.ErrNoItems
	}

	mergedItems := mergeItemsQuantities(order.Items)

	log.Printf("mergedItems: %v", mergedItems)

	// validate with stock service

	return nil
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
