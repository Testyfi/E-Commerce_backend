package payrequest

import (
	"fmt"
	"os"
)

type paymentInstrument struct {
	Type string `json:"type"`
}

type paymentRequest struct {
	MerchantID            string            `json:"merchantId"`
	MerchantTransactionID string            `json:"merchantTransactionId"`
	MerchantUserID        string            `json:"merchantUserId"`
	Amount                float64           `json:"amount"`
	RedirectURL           string            `json:"redirectUrl"`
	RedirectMode          string            `json:"redirectMode"`
	CallbackURL           string            `json:"callbackUrl"`
	MobileNumber          string            `json:"mobileNumber"`
	PaymentInstrument     paymentInstrument `json:"paymentInstrument"`
}

type FinalRequestBody struct {
	Request string `json:"request"`
}

func buildRequestPayload(payload TransactionRequest) (paymentRequest, error) {
	merchantId := os.Getenv("MERCHANT_ID")

	if merchantId == "" {
		return paymentRequest{}, fmt.Errorf("merchant id is missing")
	}

	var redirectUrl string
	var callbackUrl string

	if payload.RedirectURL == nil {
		redirectUrl = getDefaultPayRedirectUrl()
	}

	if payload.CallbackURL == nil {
		redirectUrl = getDefaultPayCallbackUrl()
	}

	paymentRequest := paymentRequest{
		MerchantID:            os.Getenv("MERCHANT_ID"),
		MerchantTransactionID: payload.TransactionID,
		MerchantUserID:        payload.UID,
		Amount:                payload.Amount,
		RedirectURL:           redirectUrl,
		RedirectMode:          "REDIRECT",
		CallbackURL:           callbackUrl,
		MobileNumber:          payload.MobileNumber,
		PaymentInstrument: paymentInstrument{
			Type: "PAY_PAGE",
		},
	}
	return paymentRequest, nil
}
