package service

import (
	"fmt"
	"strconv"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/repository"
	"github.com/google/uuid"
)

// AcquiringBank authorizes a validated payment against the acquiring bank. It
// returns whether the payment was authorized, or an error if the bank could not
// be reached or behaved unexpectedly (which is distinct from a declined
// payment). It is defined here, consumer side, so the service can be tested with
// a fake bank.
type AcquiringBank interface {
	Authorize(req models.PostPaymentRequest) (authorized bool, err error)
}

// PaymentsService holds the business logic for payments. It sits between the
// HTTP layer (api package) and the data layer (repository package), and calls
// the acquiring bank to authorize payments.
type PaymentsService struct {
	storage *repository.PaymentsRepository
	bank    AcquiringBank
}

// NewPaymentsService creates a PaymentsService backed by an in-memory repository
// and the given acquiring bank.
func NewPaymentsService(bank AcquiringBank) *PaymentsService {
	return &PaymentsService{
		storage: repository.NewPaymentsRepository(),
		bank:    bank,
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

// CreatePayment validates a payment request and, when valid, asks the acquiring
// bank to authorize it. The resulting Authorized or Declined payment is
// persisted and returned.
//
// A request that breaks a validation rule is rejected with a *RejectedError and
// never reaches the bank. A failure to reach the bank is returned as a plain
// error (neither a created payment nor a rejection).
func (s *PaymentsService) CreatePayment(req models.PostPaymentRequest) (*models.PaymentResponse, error) {
	if err := validate(req); err != nil {
		return nil, err
	}

	authorized, err := s.bank.Authorize(req)
	if err != nil {
		return nil, fmt.Errorf("authorizing payment: %w", err)
	}

	status := models.PaymentStatusDeclined
	if authorized {
		status = models.PaymentStatusAuthorized
	}

	payment := models.PaymentResponse{
		Id:                 uuid.NewString(),
		PaymentStatus:      status,
		CardNumberLastFour: lastFourDigits(req.CardNumber),
		ExpiryMonth:        req.ExpiryMonth,
		ExpiryYear:         req.ExpiryYear,
		Currency:           req.Currency,
		Amount:             req.Amount,
	}
	s.storage.AddPayment(payment)
	return &payment, nil
}

// lastFourDigits returns the last four digits of a card number as an integer.
func lastFourDigits(cardNumber string) int {
	if len(cardNumber) < 4 {
		return 0
	}
	n, _ := strconv.Atoi(cardNumber[len(cardNumber)-4:])
	return n
}
