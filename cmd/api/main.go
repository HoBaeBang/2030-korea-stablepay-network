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

	// mux는 HTTP 요청 라우터다. GET /health 같은 경로를 handler 함수로 연결한다.
	mux := http.NewServeMux()
	httpapi.RegisterHealthRoutes(mux)

	// http.Server는 API 프로세스의 주소, 라우터, 기본 timeout 설정을 가진다.
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
