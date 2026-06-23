package service

import (
	"fmt"
	"time"

	"github.com/cko-recruitment/payment-gateway-challenge-go/internal/models"
)

// RejectedError indicates that a payment request failed the gateway's validation
// rules and must not be forwarded to the acquiring bank. The HTTP layer maps it
// to a Rejected response.
type RejectedError struct {
	Reason string
}

func (e *RejectedError) Error() string {
	return e.Reason
}

// supportedCurrencies is the set of ISO currency codes the gateway accepts.
var supportedCurrencies = map[string]struct{}{
	"USD": {},
	"GBP": {},
	"EUR": {},
}

// validate checks that a payment request is well-formed. It returns a
// *RejectedError when the request breaks a rule, or nil when the payment may be
// forwarded to the acquiring bank.
func validate(req models.PostPaymentRequest) error {
	if err := validateCardNumber(req.CardNumber); err != nil {
		return err
	}
	if err := validateExpiry(req.ExpiryMonth, req.ExpiryYear); err != nil {
		return err
	}
	if err := validateCurrency(req.Currency); err != nil {
		return err
	}
	if err := validateCvv(req.Cvv); err != nil {
		return err
	}
	return nil
}

func validateCardNumber(cardNumber string) error {
	if cardNumber == "" {
		return &RejectedError{Reason: "card number is required"}
	}
	if len(cardNumber) < 14 || len(cardNumber) > 19 {
		return &RejectedError{Reason: "card number must be between 14 and 19 characters long"}
	}
	if !isAllDigits(cardNumber) {
		return &RejectedError{Reason: "card number must contain only digits"}
	}
	return nil
}

func validateExpiry(month, year int) error {
	if month < 1 || month > 12 {
		return &RejectedError{Reason: "expiry month must be between 1 and 12"}
	}
	// The card remains valid through the last day of the expiry month, i.e. up
	// to the first day of the following month.
	endOfExpiry := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	if !endOfExpiry.After(time.Now().UTC()) {
		return &RejectedError{Reason: "card expiry must be in the future"}
	}
	return nil
}

func validateCurrency(currency string) error {
	if len(currency) != 3 {
		return &RejectedError{Reason: "currency must be a 3-letter ISO code"}
	}
	if _, ok := supportedCurrencies[currency]; !ok {
		return &RejectedError{Reason: fmt.Sprintf("currency %q is not supported", currency)}
	}
	return nil
}

func validateCvv(cvv string) error {
	if len(cvv) < 3 || len(cvv) > 4 {
		return &RejectedError{Reason: "cvv must be 3 or 4 characters long"}
	}
	if !isAllDigits(cvv) {
		return &RejectedError{Reason: "cvv must contain only digits"}
	}
	return nil
}

func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
