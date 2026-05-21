package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/payment"
)

// createPaymentRequest는 HTTP JSON body를 받기 위한 요청 타입이다.
// service 계층의 CreatePaymentRequest와 분리해서 HTTP 표현과 비즈니스 입력을 구분한다.
type createPaymentRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// updatePaymentStatusRequest는 payment 상태 변경 요청 body를 받는다.
type updatePaymentStatusRequest struct {
	Status string `json:"status"`

	// transaction_hash는 ONCHAIN_DETECTED 상태에서만 필요하다.
	// 없을 수도 있으므로 *string으로 둔다.
	TransactionHash *string `json:"transaction_hash"`
}

// paymentResponse는 클라이언트에게 반환할 JSON 응답 타입이다.
type paymentResponse struct {
	ID        string `json:"id"`
	InvoiceID string `json:"invoice_id"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`

	// omitempty는 nil이면 JSON 응답에서 해당 필드를 생략한다.
	TransactionHash *string    `json:"transaction_hash,omitempty"`
	FinalizedAt     *time.Time `json:"finalized_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type paymentHandler struct {
	service *payment.Service
}

// RegisterPaymentRoutes는 payment 관련 HTTP endpoint를 등록한다.
func RegisterPaymentRoutes(mux *http.ServeMux, db *sql.DB) {
	// HTTP handler가 직접 SQL을 실행하지 않도록 repository와 service를 먼저 만든다.
	repo := payment.NewRepository(db)
	service := payment.NewService(repo)
	handler := &paymentHandler{service: service}

	mux.HandleFunc("POST /invoices/{invoiceId}/payments", handler.handleCreatePayment)
	mux.HandleFunc("PATCH /payments/{paymentId}/status", handler.handleUpdatePaymentStatus)
}

func (h *paymentHandler) handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	// URL path의 {invoiceId} 값을 꺼낸다.
	invoiceID := strings.TrimSpace(r.PathValue("invoiceId"))
	if invoiceID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invoice id is required"})
		return
	}

	// JSON body를 createPaymentRequest 구조체로 변환한다.
	var req createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	// handler는 HTTP 요청을 service 입력 타입으로 바꿔서 넘긴다.
	created, err := h.service.CreatePayment(r.Context(), payment.CreatePaymentRequest{
		InvoiceID: invoiceID,
		Amount:    req.Amount,
		Currency:  req.Currency,
	})
	if err != nil {
		if errors.Is(err, payment.ErrInvoiceNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "invoice not found"})
			return
		}

		switch err.Error() {
		case "invoice id is required", "amount must be greater than zero", "currency must be USDC":
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, toPaymentResponse(created))
}

func (h *paymentHandler) handleUpdatePaymentStatus(w http.ResponseWriter, r *http.Request) {
	// URL path의 {paymentId} 값을 꺼낸다.
	paymentID := strings.TrimSpace(r.PathValue("paymentId"))
	if paymentID == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "payment id is required"})
		return
	}

	var req updatePaymentStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json body"})
		return
	}

	// JSON의 status string을 payment.Status 타입으로 감싸서 service에 넘긴다.
	// 실제 유효한 상태인지 검증하는 책임은 service에 있다.
	updated, err := h.service.UpdatePaymentStatus(r.Context(), payment.UpdatePaymentStatusRequest{
		PaymentID:       paymentID,
		NextStatus:      payment.Status(req.Status),
		TransactionHash: req.TransactionHash,
	})
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "payment not found"})
			return
		}

		if isPaymentBadRequest(err) {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, toPaymentResponse(updated))
}

// toPaymentResponse는 domain model을 HTTP response DTO로 변환한다.
func toPaymentResponse(p *payment.Payment) paymentResponse {
	return paymentResponse{
		ID:              p.ID,
		InvoiceID:       p.InvoiceID,
		Amount:          p.Amount,
		Currency:        p.Currency,
		Status:          string(p.Status),
		TransactionHash: p.TransactionHash,
		FinalizedAt:     p.FinalizedAt,
		CreatedAt:       p.CreatedAt,
	}
}

// isPaymentBadRequest는 service error 중 400 Bad Request로 내려야 하는 에러를 구분한다.
func isPaymentBadRequest(err error) bool {
	message := err.Error()
	return strings.HasPrefix(message, "unknown payment status") ||
		strings.HasPrefix(message, "invalid payment status transition") ||
		message == "payment id is required" ||
		message == "transaction hash is required when status is ONCHAIN_DETECTED"
}
