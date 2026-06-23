package service

import (
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
)

// PaymentsService holds the business logic for payments. It sits between the
// HTTP layer (api package) and the data layer (repository package).
type PaymentsService struct {
	storage *repository.PaymentsRepository
}

// NewPaymentsService creates a PaymentsService backed by an in-memory repository.
func NewPaymentsService() *PaymentsService {
	return &PaymentsService{
		storage: repository.NewPaymentsRepository(),
	}
}

// GetPayment retrieves a payment record by its ID. It returns nil when no
// payment with the given ID exists.
func (s *PaymentsService) GetPayment(id string) *models.PaymentResponse {
	return s.storage.GetPayment(id)
}

// AddPayment persists a payment record.
func (s *PaymentsService) AddPayment(payment models.PaymentResponse) {
	s.storage.AddPayment(payment)
}
