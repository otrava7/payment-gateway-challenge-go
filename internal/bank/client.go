// Package bank provides an HTTP client for the acquiring bank that authorizes
// card payments.
package bank

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// authorizationRequest is the acquiring bank's expected request body. It differs
// from our domain model: the expiry is a single "MM/YYYY" string.
type authorizationRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	Currency   string `json:"currency"`
	Amount     int    `json:"amount"`
	Cvv        string `json:"cvv"`
}

// authorizationResponse is the acquiring bank's response body.
type authorizationResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code"`
}

// Authorize sends a (already validated) payment request to the acquiring bank
// and reports whether it was authorized. An error is returned when the bank
// cannot be reached or responds with an unexpected status, so the caller can
// distinguish a declined payment from a failed call.
func (c *Client) Authorize(req models.PostPaymentRequest) (bool, error) {
	body, err := json.Marshal(authorizationRequest{
		CardNumber: req.CardNumber,
		ExpiryDate: fmt.Sprintf("%02d/%d", req.ExpiryMonth, req.ExpiryYear),
		Currency:   req.Currency,
		Amount:     req.Amount,
		Cvv:        req.Cvv,
	})
	if err != nil {
		return false, fmt.Errorf("encoding acquiring bank request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/payments", "application/json", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("calling acquiring bank: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("acquiring bank returned unexpected status %d", resp.StatusCode)
	}

	var decoded authorizationResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return false, fmt.Errorf("decoding acquiring bank response: %w", err)
	}

	return decoded.Authorized, nil
}
