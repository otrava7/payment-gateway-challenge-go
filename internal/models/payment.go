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

// PostPaymentRequest is the payment a merchant submits to the gateway for
// processing. It is validated before being forwarded to the acquiring bank.
type PostPaymentRequest struct {
	CardNumber  string `json:"card_number"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	Currency    string `json:"currency"`
	Amount      int    `json:"amount"`
	Cvv         string `json:"cvv"`
}

// PaymentResponse is a processed payment returned to the merchant, both when the
// payment is created and when it is later retrieved by id. The full card number
// is never returned — only its last four digits.
type PaymentResponse struct {
	Id                 string        `json:"id"`
	PaymentStatus      PaymentStatus `json:"payment_status"`
	CardNumberLastFour int           `json:"card_number_last_four"`
	ExpiryMonth        int           `json:"expiry_month"`
	ExpiryYear         int           `json:"expiry_year"`
	Currency           string        `json:"currency"`
	Amount             int           `json:"amount"`
}

// PostPaymentErrorResponse is returned when a payment request is rejected before
// reaching the acquiring bank, e.g. because it failed validation.
type PostPaymentErrorResponse struct {
	PaymentStatus PaymentStatus `json:"payment_status"`
	Error         string        `json:"error"`
}

// AuthorizationRequest is the acquiring bank's expected request body. It differs
// from PostPaymentRequest: the expiry is a single "MM/YYYY" string.
type AuthorizationRequest struct {
	CardNumber string `json:"card_number"`
	ExpiryDate string `json:"expiry_date"`
	Currency   string `json:"currency"`
	Amount     int    `json:"amount"`
	Cvv        string `json:"cvv"`
}

// AuthorizationResponse is the acquiring bank's response body.
type AuthorizationResponse struct {
	Authorized        bool   `json:"authorized"`
	AuthorizationCode string `json:"authorization_code"`
}
