package main

import (
	"context"
	"log"
	"net"
	"time"

	common "github.com/salvatoreolivieri/commons"
	"github.com/salvatoreolivieri/commons/broker"
	"github.com/salvatoreolivieri/commons/discovery"
	"github.com/salvatoreolivieri/commons/discovery/consul"
	"google.golang.org/grpc"
)

var (
	serviceName = "payment"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2001")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
	amqpUser    = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass    = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost    = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort    = common.EnvString("RABBITMQ_PORT", "5672")
)

func main() {
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)

	ctx := context.Background()
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddr); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Fatal("failed to health check")
			}
			time.Sleep(time.Second * 1)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	// Message Broker Implementation
	channel, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		channel.Close()
	}()

	service := NewService()

	amqpConsumer := NewConsumer(service)
	go amqpConsumer.Listen(channel)

	// Instantiate the grpc server
	grpcServer := grpc.NewServer()

	// Listener
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	log.Println("GRPC Server started at ", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err.Error())
	}
}
