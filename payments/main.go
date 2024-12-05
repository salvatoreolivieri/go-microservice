package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	common "github.com/salvatoreolivieri/commons"
	"github.com/salvatoreolivieri/commons/broker"
	"github.com/salvatoreolivieri/commons/discovery"
	"github.com/salvatoreolivieri/commons/discovery/consul"
	"github.com/salvatoreolivieri/omsv-payments/gateway"
	stripeProcessor "github.com/salvatoreolivieri/omsv-payments/processor/stripe"
	"github.com/stripe/stripe-go/v81"
	"google.golang.org/grpc"
)

var (
	serviceName         = "payment"
	grpcAddr            = common.EnvString("GRPC_ADDR", "localhost:2001")
	consulAddr          = common.EnvString("CONSUL_ADDR", "localhost:8500")
	amqpUser            = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass            = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost            = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort            = common.EnvString("RABBITMQ_PORT", "5672")
	stripeKey           = common.EnvString("STRIPE_KEY", "sk_test_51EUwtsGYeC4mQTIrsukGOGv5iynuNR7yaZaHhP2Dv90CwtjcNHZ6NIZQB4jK2p75cn3HcBm9YohzUFIub6JnWAQ7000SWYnMjW")
	httpAddr            = common.EnvString("HTTP_ADDR", "localhost:8081")
	endpointStripSecret = common.EnvString("ENDPOINT_STRIPE_SECRET", "")
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

	// Stripe setup
	stripe.Key = stripeKey

	// Message Broker Implementation
	channel, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		channel.Close()
	}()

	// Stripe processor
	stripeProcessor := stripeProcessor.NewProcessor()
	paymentGateway := gateway.NewGateway(registry)

	service := NewService(stripeProcessor, paymentGateway)

	amqpConsumer := NewConsumer(service)
	go amqpConsumer.Listen(channel)

	// http server
	mux := http.NewServeMux()

	httpServer := NewPaymentHTTPHandler(channel)
	httpServer.registerRoutes(mux)

	go func() {
		log.Printf("Starting HTTP server at %s", httpAddr)

		if err := http.ListenAndServe(httpAddr, mux); err != nil {
			log.Fatal("failed to start http server")
		}
	}()

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
