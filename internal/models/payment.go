package models

// PaymentStatus is the outcome of a payment request.
type PaymentStatus string

const (
	// PaymentStatusAuthorized means the acquiring bank authorized the payment.
	PaymentStatusAuthorized PaymentStatus = "Authorized"
	// PaymentStatusDeclined means the acquiring bank declined the payment.
	PaymentStatusDeclined PaymentStatus = "Declined"
	// PaymentStatusRejected means the request was invalid and never reached the
	// acquiring bank, so no payment was created.
	PaymentStatusRejected PaymentStatus = "Rejected"
)

type PostPaymentRequest struct {
	CardNumber  string `json:"card_number"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	Currency    string `json:"currency"`
	Amount      int    `json:"amount"`
	Cvv         string `json:"cvv"`
}

type PaymentResponse struct {
	Id                 string        `json:"id"`
	PaymentStatus      PaymentStatus `json:"payment_status"`
	CardNumberLastFour int           `json:"card_number_last_four"`
	ExpiryMonth        int           `json:"expiry_month"`
	ExpiryYear         int           `json:"expiry_year"`
	Currency           string        `json:"currency"`
	Amount             int           `json:"amount"`
}

type PostPaymentErrorResponse struct {
	PaymentStatus PaymentStatus `json:"payment_status"`
	Error         string        `json:"error"`
}
