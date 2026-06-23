package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// newTestRouter wires a router to the payments handlers of an Api backed by the
// given (mock) service, so the HTTP layer can be exercised in isolation.
func newTestRouter(svc PaymentsService) *chi.Mux {
	a := &Api{paymentsService: svc}
	r := chi.NewRouter()
	r.Get("/api/payments/{id}", a.GetPaymentHandler())
	return r
}

func TestGetPaymentHandler(t *testing.T) {
	t.Run("PaymentFound", func(t *testing.T) {
		payment := &models.PaymentResponse{
			Id:                 "test-id",
			PaymentStatus:      "test-successful-status",
			CardNumberLastFour: 1234,
			ExpiryMonth:        10,
			ExpiryYear:         2035,
			Currency:           "GBP",
			Amount:             100,
		}
		svc := &mockPaymentsService{
			getPaymentFunc: func(id string) *models.PaymentResponse { return payment },
		}

		req := httptest.NewRequest(http.MethodGet, "/api/payments/test-id", nil)
		w := httptest.NewRecorder()
		newTestRouter(svc).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var got models.PaymentResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, *payment, got)

		// The handler should have looked the payment up by the URL id.
		assert.Equal(t, []string{"test-id"}, svc.getPaymentCalls)
	})

	t.Run("PaymentNotFound", func(t *testing.T) {
		svc := &mockPaymentsService{
			getPaymentFunc: func(id string) *models.PaymentResponse { return nil },
		}

		req := httptest.NewRequest(http.MethodGet, "/api/payments/NonExistingID", nil)
		w := httptest.NewRecorder()
		newTestRouter(svc).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, []string{"NonExistingID"}, svc.getPaymentCalls)
	})
}
