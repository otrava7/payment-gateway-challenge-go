package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// newTestRouter wires a router to the payments handlers of an Api backed by the
// given (mock) service, so the HTTP layer can be exercised in isolation.
func newTestRouter(svc PaymentsService) *chi.Mux {
	a := &Api{paymentsService: svc}
	r := chi.NewRouter()
	r.Get("/api/payments/{id}", a.GetPaymentHandler())
	r.Post("/api/payments", a.PostPaymentHandler())
	return r
}

// mockPaymentsService is a hand-written test double for the PaymentsService
// interface. It lets the HTTP handlers be tested independently of the real
// service (and its acquiring-bank HTTP client).
//
// Behaviour is configured via the function fields; calls are recorded so tests
// can assert how the handler invoked the service.
type mockPaymentsService struct {
	getPaymentFunc    func(ctx context.Context, id string) *models.PaymentResponse
	createPaymentFunc func(ctx context.Context, req models.PostPaymentRequest) (*models.PaymentResponse, error)

	getPaymentCalls    []string
	createPaymentCalls []models.PostPaymentRequest
}

func (m *mockPaymentsService) GetPayment(ctx context.Context, id string) *models.PaymentResponse {
	m.getPaymentCalls = append(m.getPaymentCalls, id)
	if m.getPaymentFunc != nil {
		return m.getPaymentFunc(ctx, id)
	}
	return nil
}

func (m *mockPaymentsService) CreatePayment(ctx context.Context, req models.PostPaymentRequest) (*models.PaymentResponse, error) {
	m.createPaymentCalls = append(m.createPaymentCalls, req)
	if m.createPaymentFunc != nil {
		return m.createPaymentFunc(ctx, req)
	}
	return nil, nil
}

func TestGetPaymentHandler(t *testing.T) {
	t.Run("Payment found", func(t *testing.T) {
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
			getPaymentFunc: func(ctx context.Context, id string) *models.PaymentResponse { return payment },
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

	t.Run("Payment not found", func(t *testing.T) {
		svc := &mockPaymentsService{
			getPaymentFunc: func(ctx context.Context, id string) *models.PaymentResponse { return nil },
		}

		req := httptest.NewRequest(http.MethodGet, "/api/payments/NonExistingID", nil)
		w := httptest.NewRecorder()
		newTestRouter(svc).ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, []string{"NonExistingID"}, svc.getPaymentCalls)
	})
}

// postPayment marshals the request and drives it through a router backed by the
// given mock service.
func postPayment(svc PaymentsService, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/payments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	newTestRouter(svc).ServeHTTP(w, req)
	return w
}

// These tests cover the handler's job in isolation: turning the service's
// outcome into an HTTP response. The validation rules themselves are exercised
// in the service package, not here.
func TestPostPaymentHandler(t *testing.T) {
	t.Run("Service rejects the payment", func(t *testing.T) {
		svc := &mockPaymentsService{
			createPaymentFunc: func(ctx context.Context, req models.PostPaymentRequest) (*models.PaymentResponse, error) {
				return nil, &service.RejectedError{Reason: "card number is required"}
			},
		}

		body, err := json.Marshal(models.PostPaymentRequest{})
		assert.NoError(t, err)

		w := postPayment(svc, body)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var got models.PostPaymentErrorResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, models.PaymentStatusRejected, got.PaymentStatus)
		assert.Equal(t, "card number is required", got.Error)

		// The handler must have delegated to the service rather than deciding itself.
		assert.Len(t, svc.createPaymentCalls, 1)
	})

	t.Run("Service authorises the payment", func(t *testing.T) {
		authorised := &models.PaymentResponse{
			Id:                 "pay-1",
			PaymentStatus:      models.PaymentStatusAuthorized,
			CardNumberLastFour: 1234,
			ExpiryMonth:        10,
			ExpiryYear:         2035,
			Currency:           "GBP",
			Amount:             100,
		}
		svc := &mockPaymentsService{
			createPaymentFunc: func(ctx context.Context, req models.PostPaymentRequest) (*models.PaymentResponse, error) {
				return authorised, nil
			},
		}

		body, err := json.Marshal(models.PostPaymentRequest{
			CardNumber:  "4111111111111234",
			ExpiryMonth: 10,
			ExpiryYear:  2035,
			Currency:    "GBP",
			Amount:      100,
			Cvv:         "123",
		})
		assert.NoError(t, err)

		w := postPayment(svc, body)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var got models.PaymentResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, *authorised, got)
	})

	t.Run("Body cannot be decoded", func(t *testing.T) {
		svc := &mockPaymentsService{}

		w := postPayment(svc, []byte("not json"))

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var got models.PostPaymentErrorResponse
		assert.NoError(t, json.NewDecoder(w.Body).Decode(&got))
		assert.Equal(t, models.PaymentStatusRejected, got.PaymentStatus)

		// A malformed body must never reach the service.
		assert.Empty(t, svc.createPaymentCalls)
	})
}
