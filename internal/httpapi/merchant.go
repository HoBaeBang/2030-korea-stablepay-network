package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/merchant"
)

type createMerchantRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type merchantResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type merchantHandler struct {
	service *merchant.Service
}

func RegisterMerchantRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := merchant.NewRepository(db)
	service := merchant.NewService(repo)
	handler := &merchantHandler{service: service}

	mux.HandleFunc("POST /merchants", handler.handlerCreateMerchant)
}

func (h *merchantHandler) handlerCreateMerchant(w http.ResponseWriter, r *http.Request) {
	var req createMerchantRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "name is required"})
		return
	}
	if req.Email == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email is required"})
		return
	}
	if !strings.Contains(req.Email, "@") {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "email is invalid"})
		return
	}

	created, err := h.service.CreateMerchant(r.Context(), merchant.CreateMerchantRequest{Name: req.Name, Email: req.Email})

	if err != nil {
		if errors.Is(err, merchant.ErrDuplicateEmail) {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "merchant email already exists"})
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, merchantResponse{
		ID:        created.ID,
		Name:      created.Name,
		Email:     created.Email,
		CreatedAt: created.CreatedAt,
	})

}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
