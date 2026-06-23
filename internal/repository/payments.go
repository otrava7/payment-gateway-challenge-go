package repository

import (
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

type PaymentsRepository struct {
	payments []models.PaymentResponse
}

func NewPaymentsRepository() *PaymentsRepository {
	return &PaymentsRepository{
		payments: []models.PaymentResponse{},
	}
}

func (ps *PaymentsRepository) GetPayment(id string) *models.PaymentResponse {
	for _, element := range ps.payments {
		if element.Id == id {
			return &element
		}
	}
	return nil
}

func (ps *PaymentsRepository) AddPayment(payment models.PaymentResponse) {
	ps.payments = append(ps.payments, payment)
}
