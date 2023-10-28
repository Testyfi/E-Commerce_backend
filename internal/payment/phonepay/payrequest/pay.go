package payrequest

import (
	"encoding/json"
	"fmt"
	"strings"
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
	xVerify, err := generatePayRequestSignature(payloadJSON)

	if err != nil {
		return TransactionResponse{}, fmt.Errorf("error: %s", err)
	}

	// make post request to get url
	client := httpClient.NewHttpClient()

	response, err := client.Post(getPayRequestEndPoint(), strings.NewReader(""),
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

	return TransactionResponse{}, nil
}
