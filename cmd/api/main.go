package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/httpapi"
	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/platform/database"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	log.Println("database connection ok")

	// mux는 multiplexer의 줄임말이다. 여러 HTTP 요청 중 경로와 method에 맞는 handler로 분배한다.
	// http.NewServeMux()는 *http.ServeMux, 즉 ServeMux 구조체의 포인터를 반환한다.
	mux := http.NewServeMux()
	httpapi.RegisterHealthRoutes(mux)
	httpapi.RegisterMerchantRoutes(mux, db)

	// &http.Server{...}는 Server 구조체 값을 만들고, 그 값의 메모리 주소를 포인터로 가져온다.
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
