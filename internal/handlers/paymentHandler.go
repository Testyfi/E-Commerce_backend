package handlers

import (
	"encoding/json"
	"net/http"
	"testify/internal/payment/phonepay/payrequest"
)

func GetPaymentRequestUrl(w http.ResponseWriter, r *http.Request) {

	payload := payrequest.TransactionRequest{
		UID:           "Anujkumarsharma123",
		Amount:        2500,
		MobileNumber:  "9517415732",
		TransactionID: "MT78505900681881048282",
	}

	transaction, err := payrequest.PayRequest(payload)

	if err != nil {
		err := json.NewEncoder(w).Encode("Something went wrong")
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		return
	}

	// Set response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}
