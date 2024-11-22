package main

import (
	"log"
	"net/http"
)

const (
	httpAddr = ":8080"
)

func main() {
	mux := http.NewServeMux()

	handler := NewHandler()
	handler.registerRoutes(mux)

	log.Printf("Starting HTTP server at %s", httpAddr)

	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal("Failed to start http server: ", err)
	}

}
