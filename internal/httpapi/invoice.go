package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/invoice"
)

type createInvoiceRequest struct {
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	ExpiresAt string `json:"expires_at"`
}

type invoiceResponse struct {
	ID         string     `json:"id"`
	MerchantID string     `json:"merchant_id"`
	Amount     int64      `json:"amount"`
	Currency   string     `json:"currency"`
	Status     string     `json:"status"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type invoiceHandler struct {
	service *invoice.Service
}

// RegisterInvoiceRoutes는 invoice 관련 HTTP endpoint를 등록한다.
func RegisterInvoiceRoutes(mux *http.ServeMux, db *sql.DB) {
	repo := invoice.NewRepository(db)
	service := invoice.NewService(repo)
	handler := &invoiceHandler{service: service}

	mux.HandleFunc("POST /merchants/{merchantId}/invoices", handler.handleCreateInvoice)
}

func (h *invoiceHandler) handleCreateInvoice(w http.ResponseWriter, r *http.Request) {
	merchantID := strings.TrimSpace(r.PathValue("merchantId"))
	if merchantID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "merchant id is required"})
		return
	}

	var req createInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	var expiresAt *time.Time
	if strings.TrimSpace(req.ExpiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "expires_at must be RFC3339"})
			return
		}
		expiresAt = &parsed
	}

	created, err := h.service.CreateInvoice(r.Context(), invoice.CreateInvoiceRequest{
		MerchantID: merchantID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		if errors.Is(err, invoice.ErrMerchantNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "merchant not found"})
			return
		}

		switch err.Error() {
		case "amount must be greater than zero", "currency is required", "currency is not supported":
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, invoiceResponse{
		ID:         created.ID,
		MerchantID: created.MerchantID,
		Amount:     created.Amount,
		Currency:   created.Currency,
		Status:     created.Status,
		ExpiresAt:  created.ExpiresAt,
		CreatedAt:  created.CreatedAt,
	})
}
