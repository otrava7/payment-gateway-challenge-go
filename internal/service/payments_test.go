package service

import (
	"errors"
	"testing"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/stretchr/testify/assert"
)

// mockBank is a test double for the AcquiringBank interface. Its behaviour is
// configured via authorizeFunc and every call is recorded.
type mockBank struct {
	authorizeFunc func(req models.PostPaymentRequest) (bool, error)
	calls         []models.PostPaymentRequest
}

func (m *mockBank) Authorize(req models.PostPaymentRequest) (bool, error) {
	m.calls = append(m.calls, req)
	if m.authorizeFunc != nil {
		return m.authorizeFunc(req)
	}
	return false, nil
}

// authorizingBank returns a mock bank that authorizes everything.
func authorizingBank() *mockBank {
	return &mockBank{authorizeFunc: func(models.PostPaymentRequest) (bool, error) { return true, nil }}
}

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

			bank := authorizingBank()
			svc := NewPaymentsService(bank)
			payment, err := svc.CreatePayment(req)

			assert.Nil(t, payment)

			var rejected *RejectedError
			assert.True(t, errors.As(err, &rejected), "expected a *RejectedError, got %v", err)

			// A rejected payment must never reach the acquiring bank.
			assert.Empty(t, bank.calls)
		})
	}

	t.Run("non-positive amount is not rejected", func(t *testing.T) {
		// The gateway does not constrain the sign of the amount; it is left to
		// the acquiring bank to accept or decline it.
		svc := NewPaymentsService(authorizingBank())

		req := validRequest()
		req.Amount = -100

		payment, err := svc.CreatePayment(req)

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, -100, payment.Amount)
	})
}

func TestCreatePaymentAuthorization(t *testing.T) {
	t.Run("authorized payment is stored", func(t *testing.T) {
		bank := authorizingBank()
		svc := NewPaymentsService(bank)

		payment, err := svc.CreatePayment(validRequest())

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.NotEmpty(t, payment.Id)
		assert.Equal(t, models.PaymentStatusAuthorized, payment.PaymentStatus)
		assert.Equal(t, 1234, payment.CardNumberLastFour)
		assert.Equal(t, "GBP", payment.Currency)
		assert.Equal(t, 100, payment.Amount)

		// The validated request was forwarded to the bank, and the resulting
		// payment is retrievable by its generated id.
		assert.Len(t, bank.calls, 1)
		assert.Equal(t, payment, svc.GetPayment(payment.Id))
	})

	t.Run("declined payment is stored", func(t *testing.T) {
		bank := &mockBank{authorizeFunc: func(models.PostPaymentRequest) (bool, error) { return false, nil }}
		svc := NewPaymentsService(bank)

		payment, err := svc.CreatePayment(validRequest())

		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, models.PaymentStatusDeclined, payment.PaymentStatus)

		// A declined payment is still a created payment and must be retrievable.
		assert.Equal(t, payment, svc.GetPayment(payment.Id))
	})

	t.Run("bank failure is not a rejection and is not stored", func(t *testing.T) {
		bankErr := errors.New("bank unavailable")
		bank := &mockBank{authorizeFunc: func(models.PostPaymentRequest) (bool, error) { return false, bankErr }}
		svc := NewPaymentsService(bank)

		payment, err := svc.CreatePayment(validRequest())

		assert.Nil(t, payment)
		assert.ErrorIs(t, err, bankErr)

		// A bank failure is distinct from a validation rejection.
		var rejected *RejectedError
		assert.False(t, errors.As(err, &rejected))
	})
}
