package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/service"
	"github.com/go-chi/chi/v5"
)

// GetPaymentHandler returns an http.HandlerFunc that handles Payments GET requests.
// It retrieves a payment record by its ID from the payments service.
// The ID is expected to be part of the URL.
func (a *Api) GetPaymentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		payment := a.paymentsService.GetPayment(r.Context(), id)

		if payment == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		writeJSON(w, http.StatusOK, payment)
	}
}

// PostPaymentHandler returns an http.HandlerFunc that handles Payments POST requests.
// It decodes the request and delegates processing to the payments service,
// mapping the service's outcome to an HTTP response. The payment rules
// themselves live in the service, not here.
func (a *Api) PostPaymentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.PostPaymentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// A body that cannot be decoded is a transport error, not a payment
			// rule, so the API layer handles it directly.
			writeJSON(w, http.StatusBadRequest, models.PostPaymentErrorResponse{
				PaymentStatus: models.PaymentStatusRejected,
				Error:         "invalid request body",
			})
			return
		}

		payment, err := a.paymentsService.CreatePayment(r.Context(), req)
		if err != nil {
			var rejected *service.RejectedError
			if errors.As(err, &rejected) {
				slog.WarnContext(r.Context(), "payment rejected", "reason", rejected.Reason)
				writeJSON(w, http.StatusBadRequest, models.PostPaymentErrorResponse{
					PaymentStatus: models.PaymentStatusRejected,
					Error:         rejected.Reason,
				})
				return
			}
			// An unexpected failure (e.g. the acquiring bank is unreachable): log
			// the cause for operators, but don't leak internals to the caller.
			slog.ErrorContext(r.Context(), "payment processing failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, models.PostPaymentErrorResponse{
				PaymentStatus: models.PaymentStatusRejected,
				Error:         "could not process payment",
			})
			return
		}

		writeJSON(w, http.StatusCreated, payment)
	}
}

// writeJSON writes the given status code and JSON-encoded body to the response.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// The header (including status) is already sent, so an encode failure here
	// can only be logged, not turned into a different status code.
	_ = json.NewEncoder(w).Encode(body)
}
