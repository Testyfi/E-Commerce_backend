package handlers

import (
	"net/http"
	"testify/internal/payment/phonepay/payrequest"
	httpClient "testify/internal/utility/http"
)

func GetPaymentRequestUrl(w http.ResponseWriter, r *http.Request) {

	payload := payrequest.TransactionRequest{
		UID:           "Anujkumarsharma123",
		Amount:        2829,
		MobileNumber:  "9517415732",
		TransactionID: "MT78505900681881048298",
	}

	transaction, err := payrequest.PayRequest(payload)

	if err != nil {
		httpClient.RespondError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	httpClient.RespondSuccess(w, transaction)
}
