package bank

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func validRequest() models.PostPaymentRequest {
	return models.PostPaymentRequest{
		CardNumber:  "4111111111111234",
		ExpiryMonth: 4,
		ExpiryYear:  2025,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         "123",
	}
}

func TestClientAuthorize(t *testing.T) {
	t.Run("maps the request and returns authorized", func(t *testing.T) {
		var gotBody map[string]any
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/payments", r.URL.Path)
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"authorized":         true,
				"authorization_code": "0bb07405-6d44-4b50-a14f-7ae0beff13ad",
			})
		}))
		defer server.Close()

		authorized, err := NewClient(server.URL).Authorize(validRequest())

		assert.NoError(t, err)
		assert.True(t, authorized)

		// The domain model's month/year is mapped to the bank's "MM/YYYY" field.
		assert.Equal(t, "4111111111111234", gotBody["card_number"])
		assert.Equal(t, "04/2025", gotBody["expiry_date"])
		assert.Equal(t, "GBP", gotBody["currency"])
		assert.Equal(t, float64(100), gotBody["amount"])
		assert.Equal(t, "123", gotBody["cvv"])
	})

	t.Run("returns declined", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"authorized": false, "authorization_code": ""})
		}))
		defer server.Close()

		authorized, err := NewClient(server.URL).Authorize(validRequest())

		assert.NoError(t, err)
		assert.False(t, authorized)
	})

	t.Run("non-200 status is an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		authorized, err := NewClient(server.URL).Authorize(validRequest())

		assert.Error(t, err)
		assert.False(t, authorized)
	})
}
