package payrequest

type redirectInfo struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

type instrumentResponse struct {
	Type         string       `json:"type"`
	RedirectInfo redirectInfo `json:"redirectInfo"`
}

type Data struct {
	MerchantID            string             `json:"merchantId"`
	MerchantTransactionID string             `json:"merchantTransactionId"`
	InstrumentResponse    instrumentResponse `json:"instrumentResponse"`
}

type paymentResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}
