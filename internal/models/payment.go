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
	// Card number: 14–19 characters, digits only.
	CardNumber string `json:"card_number" validate:"required" example:"2222405343248877"`
	// Expiry month, 1–12. The month/year combination must be in the future.
	ExpiryMonth int `json:"expiry_month" validate:"required" example:"4"`
	// Expiry year. The month/year combination must be in the future.
	ExpiryYear int `json:"expiry_year" validate:"required" example:"2030"`
	// ISO currency code; must be one of USD, GBP, EUR.
	Currency string `json:"currency" validate:"required" example:"GBP"`
	// Amount in the minor currency unit, e.g. 1050 = £10.50.
	Amount int `json:"amount" validate:"required" example:"1050"`
	// Card verification value: 3–4 characters, digits only.
	Cvv string `json:"cvv" validate:"required" example:"123"`
}

// PaymentResponse is a processed payment returned to the merchant, both when the
// payment is created and when it is later retrieved by id. The full card number
// is never returned — only its last four digits.
type PaymentResponse struct {
	// Id is the gateway-assigned payment id (a GUID), used to retrieve the payment.
	Id string `json:"id"`
	// PaymentStatus is the outcome of a created payment: Authorized or Declined.
	PaymentStatus PaymentStatus `json:"payment_status" swaggertype:"string" enums:"Authorized,Declined"`
	// CardNumberLastFour is the last four digits of the card used.
	CardNumberLastFour int `json:"card_number_last_four"`
	// ExpiryMonth of the card, 1–12.
	ExpiryMonth int `json:"expiry_month"`
	// ExpiryYear of the card.
	ExpiryYear int `json:"expiry_year"`
	// Currency is the ISO currency code of the payment.
	Currency string `json:"currency"`
	// Amount in the minor currency unit, e.g. 1050 = £10.50.
	Amount int `json:"amount"`
}

// PostPaymentErrorResponse is returned when a payment request is rejected before
// reaching the acquiring bank, e.g. because it failed validation.
type PostPaymentErrorResponse struct {
	// PaymentStatus is always "Rejected" for an error response.
	PaymentStatus PaymentStatus `json:"payment_status" swaggertype:"string" enums:"Rejected"`
	// Error is a human-readable description of why the request could not be processed.
	Error string `json:"error"`
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
