package payrequest

import (
	"bytes"
	"encoding/json"
	"fmt"
	httpClient "testify/internal/utility/http"
)

func PayRequest(payload TransactionRequest) (TransactionResponse, error) {

	// build request
	requestPayload, err := buildRequestPayload(payload)
	if err != nil {
		return TransactionResponse{}, fmt.Errorf("error: %s", err)
	}

	payloadJSON, err := json.Marshal(requestPayload)

	if err != nil {
		return TransactionResponse{}, fmt.Errorf("error: %s", err)
	}

	// generate signature
	finalRequest, xVerify, err := generatePayRequestSignature(payloadJSON)

	if err != nil {
		return TransactionResponse{}, fmt.Errorf("error: %s", err)
	}

	reqBody, _ := json.Marshal(finalRequest)

	// make post request to get url
	client := httpClient.NewHttpClient()

	response, err := client.Post(getPayRequestEndPoint(), bytes.NewReader(reqBody),
		httpClient.WithHeader("X-VERIFY", xVerify))

	if err != nil {
		return TransactionResponse{}, fmt.Errorf("error: %s", err)
	}

	// get request url
	var paymentResp paymentResponse
	err = json.Unmarshal([]byte(response), &paymentResp)
	if err != nil {
		return TransactionResponse{}, fmt.Errorf("failed to unmarshal response: %s", err)
	}

	return TransactionResponse{
		RedirectUrl: paymentResp.Data.InstrumentResponse.RedirectInfo.URL,
	}, nil
}
