package gateway

import (
	"context"
	"log"

	pb "github.com/salvatoreolivieri/commons/api"
	"github.com/salvatoreolivieri/commons/discovery"
)

type gateway struct {
	registry discovery.Registry
}

func NewGRPCGateway(registry discovery.Registry) *gateway {
	return &gateway{registry}
}

func (g *gateway) CreateOrder(ctx context.Context, payload *pb.CreateOrderRequest) (*pb.Order, error) {
	conn, err := discovery.ServiceConnection(ctx, "orders", g.registry)
	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	client := pb.NewOrderServiceClient(conn)

	return client.CreateOrder(ctx, &pb.CreateOrderRequest{
		CustomerID: payload.CustomerID,
		Items:      payload.Items,
	})

}
