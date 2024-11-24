package main

import (
	"context"
	"log"
	"net"

	common "github.com/salvatoreolivieri/commons"
	"google.golang.org/grpc"
)

var (
	grpcAddr = common.EnvString("GRPC_ADDR", "localhost:2000")
)

func main() {

	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	store := NewStore()
	service := NewService(store)

	NewGRPCHandler(grpcServer, service)

	service.CreateOrder(context.Background())

	log.Println("GRPC Server started at ", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err.Error())
	}
}
