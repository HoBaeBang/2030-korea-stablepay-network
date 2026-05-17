package httpapi

import (
	"encoding/json"
	"net/http"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func RegisterHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", handleHealth)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:  "ok",
		Service: "stablepay-api",
	})
}
