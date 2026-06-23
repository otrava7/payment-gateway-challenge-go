package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetPaymentHandler returns an http.HandlerFunc that handles Payments GET requests.
// It retrieves a payment record by its ID from the payments service.
// The ID is expected to be part of the URL.
func (a *Api) GetPaymentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		payment := a.paymentsService.GetPayment(id)

		if payment != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(payment); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// PostPaymentHandler returns an http.HandlerFunc that handles Payments POST requests.
func (a *Api) PostPaymentHandler() http.HandlerFunc {
	//TODO
	return nil
}
