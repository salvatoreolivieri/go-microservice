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
	serviceName = "orders"
	grpcAddr    = common.EnvString("GRPC_ADDR", "localhost:2000")
	consulAddr  = common.EnvString("CONSUL_ADDR", "localhost:8500")
	amqpUser    = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass    = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost    = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort    = common.EnvString("RABBITMQ_PORT", "5672")
	jaegerAddr  = common.EnvString("JAEGER_ADDR", "localhost:4318")
)

func main() {
	err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr)
	if err != nil {
		log.Fatal("failed to set up global tracer")
	}

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

	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	store := NewStore()
	service := NewService(store)
	serviceWithTelemetry := NewTelemetryMiddleware(service)

	NewGRPCHandler(grpcServer, serviceWithTelemetry, channel)

	consumer := NewConsumer(serviceWithTelemetry)
	go consumer.Listen(channel)

	log.Println("GRPC Server started at ", grpcAddr)

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err.Error())
	}
}
