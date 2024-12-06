package main

import (
	"context"
	"fmt"
	"net"
	"time"

	common "github.com/salvatoreolivieri/commons"
	"github.com/salvatoreolivieri/commons/broker"
	"github.com/salvatoreolivieri/commons/discovery"
	"github.com/salvatoreolivieri/commons/discovery/consul"
	"github.com/salvatoreolivieri/omsv-orders/gateway"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
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
	mongoUser   = common.EnvString("MONGO_DB_USER", "root")
	mongoPass   = common.EnvString("MONGO_DB_PASS", "example")
	mongoAddr   = common.EnvString("MONGO_DB_HOST", "localhost:27017")
	jaegerAddr  = common.EnvString("JAEGER_ADDR", "localhost:4318")
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr)
	if err != nil {
		logger.Fatal("coul not set global tracer", zap.Error(err))
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
				logger.Error("failed to health check", zap.Error(err))
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

	// mongo db conn
	uri := fmt.Sprintf("mongodb://%s:%s@%s", mongoUser, mongoPass, mongoAddr)
	mongoClient, err := connectToMongoDB(uri)
	if err != nil {
		logger.Fatal("failed to connect to mongo db", zap.Error(err))
	}

	grpcServer := grpc.NewServer()

	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Fatal("failer to listen", zap.Error(err))
	}
	defer listener.Close()

	gateway := gateway.NewGateway(registry)

	store := NewStore(mongoClient)
	service := NewService(store, gateway)
	serviceWithTelemetry := NewTelemetryMiddleware(service)
	serviceWithLogging := NewLoggingMiddleware(serviceWithTelemetry)

	NewGRPCHandler(grpcServer, serviceWithLogging, channel)

	consumer := NewConsumer(serviceWithTelemetry)
	go consumer.Listen(channel)

	logger.Info("GRPC Server started", zap.String("port", grpcAddr))

	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatal("failed to serve", zap.Error(err))
	}
}

func connectToMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	return client, err
}
