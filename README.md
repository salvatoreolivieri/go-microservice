# Order Management System with Microservices in Go

This project implements an Order Management System using a microservices architecture in Go. It consists of five distinct microservices that communicate and work together to complete a user's order.

![Microservice Architecture](/microservice-architecture.png "Microservice Architecture")

### Microservice overview:

**Orders Service**

- Validate Order Details -> Talk with stock service
- CRUD of orders
- Initiates the Payment Flow -> by sending an event

**Stock Service**

- Handle Stock
- Validates order quantities
- Return items as menu

**Menu Service**

- Store items as menu

**Payment Service**

- Initiates a payment with a 3rd part provider
- Produces an order paid/cancelled event to orders, stock and kitchen

**Kitchen Service**

- Long running process of a "Simulated kitchen staff"

### Tech Overview:

- **Go 1.23+** - The primary programming language.
- **Golang Cosmtrek/Air** - For hot-reloading during development.
- **gRPC** - Communication protocol between services.
- **RabbitMQ** - Message broker for handling event-driven communication.
- **Docker & Docker Compose** - For containerization and orchestration of microservices.
- **MongoDB** - Storage layer for persisting data.
- **Jaeger** - For distributed tracing and performance monitoring.
- **HashiCorp Consul** - Service discovery tool.
- **Stripe** - Payment processor for handling transactions.
