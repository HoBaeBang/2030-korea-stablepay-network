package httpapi

import (
	"encoding/json"
	"net/http"
)

// healthResponse는 GET /health 요청에 반환할 JSON 응답 본문이다.
type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// RegisterHealthRoutes는 health check endpoint를 공용 HTTP 라우터에 등록한다.
func RegisterHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", handleHealth)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode는 struct 값을 JSON으로 변환해 HTTP 응답 body에 기록한다.
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:  "ok",
		Service: "stablepay-api",
	})
}
