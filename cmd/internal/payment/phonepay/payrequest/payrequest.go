package payrequest

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"testify/internal/payment/phonepay"
)

func getPayRequestEndPoint() string {
	return phonepay.GetBaseEndpoint() + phonepay.PayEndPoint
}

func getDefaultPayRedirectUrl() string {
	return "https://webhook.site/redirect-url"
}

func getDefaultPayCallbackUrl() string {
	return "https://webhook.site/369b0c9d-2c52-4db4-8a26-9a942c882990"
}

func generatePayRequestSignature(payload []byte) (finalRequestBody, string, error) {
	saltKey := os.Getenv("SALT_KEY")
	saltIndex := os.Getenv("SALT_INDEX")

	if saltKey == "" {
		return finalRequestBody{}, "", fmt.Errorf("salt key is not provided")
	}

	if saltIndex == "" {
		return finalRequestBody{}, "", fmt.Errorf("salt index is not provided")
	}

	encodedPayload := base64.StdEncoding.EncodeToString(payload)
	toHash := encodedPayload + phonepay.PayEndPoint + saltKey
	hash := sha256.Sum256([]byte(toHash))
	return finalRequestBody{Request: encodedPayload}, fmt.Sprintf("%x###%s", hash, saltIndex), nil
}
