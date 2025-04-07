package payrequest

type TransactionRequest struct {
	UID           string
	Amount        float64
	MobileNumber  string
	TransactionID string
	RedirectURL   *string
	CallbackURL   *string
}

type TransactionResponse struct {
	RedirectUrl string `json:"payment_url"`
}
