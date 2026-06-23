package models

type PostPaymentRequest struct {
	CardNumber  string `json:"card_number"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	Currency    string `json:"currency"`
	Amount      int    `json:"amount"`
	Cvv         string `json:"cvv"`
}

type PaymentResponse struct {
	Id                 string `json:"id"`
	PaymentStatus      string `json:"payment_status"`
	CardNumberLastFour int    `json:"card_number_last_four"`
	ExpiryMonth        int    `json:"expiry_month"`
	ExpiryYear         int    `json:"expiry_year"`
	Currency           string `json:"currency"`
	Amount             int    `json:"amount"`
}

type PostPaymentErrorResponse struct {
	PaymentStatus string `json:"payment_status"`
	Error         string `json:"error"`
}
