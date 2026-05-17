package httpapi

import (
	"encoding/json"
	"net/http"
)

// healthResponse is the JSON body returned by GET /health.
type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// RegisterHealthRoutes attaches health-check endpoints to the shared HTTP router.
func RegisterHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", handleHealth)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode writes the response struct as JSON to the HTTP response body.
	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:  "ok",
		Service: "stablepay-api",
	})
}
