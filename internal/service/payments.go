package service

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

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

// CreatePayment validates a payment request and, when valid, processes and
// persists it. A request that breaks a validation rule is rejected with a
// *RejectedError and never reaches the acquiring bank.
//
// TODO: forward valid requests to the acquiring bank to determine whether the
// payment is Authorized or Declined. Until that HTTP client exists the payment
// is recorded as Authorized.
func (s *PaymentsService) CreatePayment(req models.PostPaymentRequest) (*models.PaymentResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}

	payment := models.PaymentResponse{
		Id:                 newID(),
		PaymentStatus:      "Authorized",
		CardNumberLastFour: lastFourDigits(req.CardNumber),
		ExpiryMonth:        req.ExpiryMonth,
		ExpiryYear:         req.ExpiryYear,
		Currency:           req.Currency,
		Amount:             req.Amount,
	}
	s.storage.AddPayment(payment)
	return &payment, nil
}

// newID returns a random hex identifier for a payment.
func newID() string {
	b := make([]byte, 16)
	// crypto/rand.Read never returns an error on the supported platforms.
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// lastFourDigits returns the last four digits of a card number as an integer.
func lastFourDigits(cardNumber string) int {
	if len(cardNumber) < 4 {
		return 0
	}
	n, _ := strconv.Atoi(cardNumber[len(cardNumber)-4:])
	return n
}
