package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/httpapi"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// mux is the HTTP request router. It maps paths like GET /health to handler functions.
	mux := http.NewServeMux()
	httpapi.RegisterHealthRoutes(mux)

	// http.Server owns the network address, router, and basic timeout settings for the API process.
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("stablepay api listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
