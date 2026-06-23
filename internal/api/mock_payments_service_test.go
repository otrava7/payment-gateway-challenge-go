package api

import "github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"

// mockPaymentsService is a hand-written test double for the PaymentsService
// interface. It lets the HTTP handlers be tested independently of the real
// service (and, later, its acquiring-bank HTTP client).
//
// Behaviour is configured via the function fields; calls are recorded so tests
// can assert how the handler invoked the service.
type mockPaymentsService struct {
	getPaymentFunc    func(id string) *models.PaymentResponse
	createPaymentFunc func(req models.PostPaymentRequest) (*models.PaymentResponse, error)

	getPaymentCalls    []string
	createPaymentCalls []models.PostPaymentRequest
}

func (m *mockPaymentsService) GetPayment(id string) *models.PaymentResponse {
	m.getPaymentCalls = append(m.getPaymentCalls, id)
	if m.getPaymentFunc != nil {
		return m.getPaymentFunc(id)
	}
	return nil
}

func (m *mockPaymentsService) CreatePayment(req models.PostPaymentRequest) (*models.PaymentResponse, error) {
	m.createPaymentCalls = append(m.createPaymentCalls, req)
	if m.createPaymentFunc != nil {
		return m.createPaymentFunc(req)
	}
	return nil, nil
}
