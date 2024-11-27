package main

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/salvatoreolivieri/commons/api"

	"github.com/salvatoreolivieri/commons/broker"
)

type consumer struct {
	service PaymentsService
}

func NewConsumer(service PaymentsService) *consumer {
	return &consumer{service}
}

func (c *consumer) Listen(channel *amqp.Channel) {

	que, err := channel.QueueDeclare(broker.OrderCreatedEvent, true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	messages, err := channel.Consume(que.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	var forever chan struct{}

	go func() {
		for delivery := range messages {
			log.Printf("Received message: %v", delivery.Body)

			order := &pb.Order{}
			if err := json.Unmarshal(delivery.Body, order); err != nil {
				log.Printf("ailed to unmarshal order: %v", err)
				continue
			}

			paymentLink, err := c.service.CreatePayment(context.Background(), order)
			if err != nil {
				log.Printf("failed to create payment: %v", err)
				continue
			}

			log.Printf("payment link created %s", paymentLink)
		}
	}()

	<-forever
}
