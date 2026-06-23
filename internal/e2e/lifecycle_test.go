//go:build e2e

// Package e2e holds end-to-end tests that drive the whole stack: a real HTTP
// request hits the wired-up router, which flows through the handler, the
// service, a real HTTP call to the mountebank bank simulator (see
// docker-compose.yml), and the in-memory storage — then the payment is
// retrieved back over HTTP.
//
// These tests are excluded from the normal `go test ./...` run by the build tag
// and are executed with `go test -tags=e2e ./...` once the simulator is up.
package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/api"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newServer wires up the full Api (with the real acquiring-bank client) and
// serves it over real HTTP for the duration of the test.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	if u := os.Getenv("ACQUIRING_BANK_URL"); u != "" {
		t.Setenv("ACQUIRING_BANK_URL", u)
	} else {
		t.Setenv("ACQUIRING_BANK_URL", "http://localhost:8080")
	}

	server := httptest.NewServer(api.New().Handler())
	t.Cleanup(server.Close)
	return server
}

func postPayment(t *testing.T, server *httptest.Server, req models.PostPaymentRequest) *http.Response {
	t.Helper()
	body, err := json.Marshal(req)
	require.NoError(t, err)

	resp, err := http.Post(server.URL+"/api/payments", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	return resp
}

func validRequest(cardNumber string) models.PostPaymentRequest {
	return models.PostPaymentRequest{
		CardNumber:  cardNumber,
		ExpiryMonth: 4,
		ExpiryYear:  2030,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         "123",
	}
}

func TestPaymentLifecycleE2E(t *testing.T) {
	t.Run("authorized payment is created and retrievable", func(t *testing.T) {
		server := newServer(t)

		// Card ending in an odd digit -> the simulator authorizes it.
		resp := postPayment(t, server, validRequest("2222405343248877"))
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var created models.PaymentResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
		assert.NotEmpty(t, created.Id)
		assert.Equal(t, models.PaymentStatusAuthorized, created.PaymentStatus)
		assert.Equal(t, 8877, created.CardNumberLastFour)

		// Retrieve the same payment back over HTTP.
		getResp, err := http.Get(server.URL + "/api/payments/" + created.Id)
		require.NoError(t, err)
		defer getResp.Body.Close()

		require.Equal(t, http.StatusOK, getResp.StatusCode)

		var fetched models.PaymentResponse
		require.NoError(t, json.NewDecoder(getResp.Body).Decode(&fetched))
		assert.Equal(t, created, fetched)
	})

	t.Run("declined payment is created and retrievable", func(t *testing.T) {
		server := newServer(t)

		// Card ending in an even digit -> the simulator declines it.
		resp := postPayment(t, server, validRequest("2222405343248112"))
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var created models.PaymentResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
		assert.Equal(t, models.PaymentStatusDeclined, created.PaymentStatus)

		getResp, err := http.Get(server.URL + "/api/payments/" + created.Id)
		require.NoError(t, err)
		defer getResp.Body.Close()

		require.Equal(t, http.StatusOK, getResp.StatusCode)
	})

	t.Run("invalid request is rejected and never stored", func(t *testing.T) {
		server := newServer(t)

		req := validRequest("2222405343248877")
		req.CardNumber = "" // invalid -> rejected before the bank is called

		resp := postPayment(t, server, req)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var rejected models.PostPaymentErrorResponse
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&rejected))
		assert.Equal(t, models.PaymentStatusRejected, rejected.PaymentStatus)
	})

	t.Run("unknown payment id returns not found", func(t *testing.T) {
		server := newServer(t)

		getResp, err := http.Get(server.URL + "/api/payments/does-not-exist")
		require.NoError(t, err)
		defer getResp.Body.Close()

		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})
}
