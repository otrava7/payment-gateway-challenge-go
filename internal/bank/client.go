// Package bank provides an HTTP client for the acquiring bank that authorizes
// card payments.
package bank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

// Client talks to the acquiring bank over HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a Client that sends authorization requests to baseURL
// (e.g. "http://localhost:8080").
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Authorize sends a (already validated) payment request to the acquiring bank
// and reports whether it was authorized. An error is returned when the bank
// cannot be reached or responds with an unexpected status, so the caller can
// distinguish a declined payment from a failed call.
func (c *Client) Authorize(ctx context.Context, req models.PostPaymentRequest) (bool, error) {
	body, err := json.Marshal(models.AuthorizationRequest{
		CardNumber: req.CardNumber,
		ExpiryDate: fmt.Sprintf("%02d/%d", req.ExpiryMonth, req.ExpiryYear),
		Currency:   req.Currency,
		Amount:     req.Amount,
		Cvv:        req.Cvv,
	})
	if err != nil {
		return false, fmt.Errorf("encoding acquiring bank request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/payments", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("building acquiring bank request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false, fmt.Errorf("calling acquiring bank: %w", err)
	}
	defer resp.Body.Close()

	// Latency and status only — the request body carries the card number and
	// CVV, which must never be logged (PCI-DSS).
	slog.InfoContext(ctx, "acquiring bank responded",
		"status", resp.StatusCode,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("acquiring bank returned unexpected status %d", resp.StatusCode)
	}

	var decoded models.AuthorizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return false, fmt.Errorf("decoding acquiring bank response: %w", err)
	}

	return decoded.Authorized, nil
}
