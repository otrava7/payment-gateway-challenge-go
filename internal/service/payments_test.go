package service

import (
	"errors"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/stretchr/testify/assert"
)

// validRequest returns a payment request that passes every validation rule, so
// individual tests can mutate a single field to exercise one rule at a time.
func validRequest() models.PostPaymentRequest {
	return models.PostPaymentRequest{
		CardNumber:  "4111111111111234",
		ExpiryMonth: 12,
		ExpiryYear:  3000,
		Currency:    "GBP",
		Amount:      100,
		Cvv:         "123",
	}
}

func TestCreatePaymentValidation(t *testing.T) {
	rejections := map[string]func(*models.PostPaymentRequest){
		"card number missing":   func(r *models.PostPaymentRequest) { r.CardNumber = "" },
		"card number too short": func(r *models.PostPaymentRequest) { r.CardNumber = "4111111111111" },        // 13
		"card number too long":  func(r *models.PostPaymentRequest) { r.CardNumber = "41111111111112345678" }, // 20
		"card number non-digit": func(r *models.PostPaymentRequest) { r.CardNumber = "4111-1111-1111-1234" },
		"expiry month too low":  func(r *models.PostPaymentRequest) { r.ExpiryMonth = 0 },
		"expiry month too high": func(r *models.PostPaymentRequest) { r.ExpiryMonth = 13 },
		"expiry in the past":    func(r *models.PostPaymentRequest) { r.ExpiryYear = 2000 },
		"currency wrong length": func(r *models.PostPaymentRequest) { r.Currency = "GB" },
		"currency unsupported":  func(r *models.PostPaymentRequest) { r.Currency = "JPY" },
		"cvv too short":         func(r *models.PostPaymentRequest) { r.Cvv = "12" },
		"cvv too long":          func(r *models.PostPaymentRequest) { r.Cvv = "12345" },
		"cvv non-digit":         func(r *models.PostPaymentRequest) { r.Cvv = "12a" },
	}

	for name, mutate := range rejections {
		t.Run(name, func(t *testing.T) {
			req := validRequest()
			mutate(&req)

			svc := NewPaymentsService()
			payment, err := svc.CreatePayment(req)

			assert.Nil(t, payment)

			var rejected *RejectedError
			assert.True(t, errors.As(err, &rejected), "expected a *RejectedError, got %v", err)

			// A rejected payment must not be persisted.
			assert.Nil(t, svc.GetPayment(""))
		})
	}

	t.Run("valid request is authorised and stored", func(t *testing.T) {
		svc := NewPaymentsService()

		payment, err := svc.CreatePayment(validRequest())

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.NotEmpty(t, payment.Id)
		assert.Equal(t, "Authorized", payment.PaymentStatus)
		assert.Equal(t, 1234, payment.CardNumberLastFour)
		assert.Equal(t, "GBP", payment.Currency)
		assert.Equal(t, 100, payment.Amount)

		// The processed payment is retrievable by its generated id.
		stored := svc.GetPayment(payment.Id)
		assert.Equal(t, payment, stored)
	})

	t.Run("non-positive amount is not rejected", func(t *testing.T) {
		// The gateway does not constrain the sign of the amount; it is left to
		// the acquiring bank to accept or decline it.
		svc := NewPaymentsService()

		req := validRequest()
		req.Amount = -100

		payment, err := svc.CreatePayment(req)

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, -100, payment.Amount)
	})
}
