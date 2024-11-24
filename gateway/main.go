package main

import (
	"log"
	"net/http"

	common "github.com/salvatoreolivieri/commons"
	pb "github.com/salvatoreolivieri/commons/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	httpAddr         = common.EnvString("HTTP_ADDR", ":8080")
	orderServiceAddr = "localhost:2000"
)

func main() {
	conn, err := grpc.NewClient(orderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to dial server: %v", err)
	}
	defer conn.Close()

	log.Printf("Dialing orders service at ", orderServiceAddr)

	client := pb.NewOrderServiceClient(conn)

	mux := http.NewServeMux()

	handler := NewHandler(client)
	handler.registerRoutes(mux)

	log.Printf("Starting HTTP server at %s", httpAddr)

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal("Failed to start http server: ", err)
	}

}
