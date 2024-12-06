package main

import (
	"context"

	pb "github.com/salvatoreolivieri/commons/api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	dbName         = "orders"
	collectionName = "orders"
)

var orders = make([]*pb.Order, 0)

type store struct {
	db *mongo.Client
}

func NewStore(db *mongo.Client) *store {
	return &store{db}
}

func (s *store) Create(ctx context.Context, order Order) (primitive.ObjectID, error) {
	col := s.db.Database(dbName).Collection(collectionName)
	newOrder, err := col.InsertOne(ctx, order)
	if err != nil {
		return primitive.ObjectID{}, err
	}
	id := newOrder.InsertedID.(primitive.ObjectID)

	return id, nil
}

func (s *store) Get(ctx context.Context, orderID, customerID string) (*Order, error) {
	col := s.db.Database(dbName).Collection(collectionName)

	oID, _ := primitive.ObjectIDFromHex(orderID)

	var order Order
	err := col.FindOne(ctx, bson.M{
		"_id":         oID,
		"customer_id": customerID,
	}).Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *store) Update(ctx context.Context, orderID string, newOrder *pb.Order) error {
	col := s.db.Database(dbName).Collection(collectionName)

	oID, _ := primitive.ObjectIDFromHex(orderID)

	_, err := col.UpdateOne(ctx,
		bson.M{"_id": oID},
		bson.M{"$set": bson.M{
			"paymentLink": newOrder.PaymentLink,
			"status":      newOrder.Status,
		}})

	return err
}
