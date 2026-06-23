package api

import "github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"

// mockPaymentsService is a hand-written test double for the PaymentsService
// interface. It lets the HTTP handlers be tested independently of the real
// service (and, later, its acquiring-bank HTTP client).
//
// Behaviour is configured via the function fields; calls are recorded so tests
// can assert how the handler invoked the service.
type mockPaymentsService struct {
	getPaymentFunc func(id string) *models.PaymentResponse

	getPaymentCalls []string
}

func (m *mockPaymentsService) GetPayment(id string) *models.PaymentResponse {
	m.getPaymentCalls = append(m.getPaymentCalls, id)
	if m.getPaymentFunc != nil {
		return m.getPaymentFunc(id)
	}
	return nil
}
