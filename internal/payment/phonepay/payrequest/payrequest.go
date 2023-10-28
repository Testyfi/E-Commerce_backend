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
	return "https://webhook.site/callback-url"
}

func generatePayRequestSignature(payload []byte) (string, error) {
	saltKey := os.Getenv("SALT_KEY")
	saltIndex := os.Getenv("SALT_INDEX")

	if saltKey == "" {
		return "", fmt.Errorf("salt key is not provided")
	}

	if saltIndex == "" {
		return "", fmt.Errorf("salt index is not provided")
	}

	encodedPayload := base64.StdEncoding.EncodeToString(payload)
	toHash := encodedPayload + phonepay.PayEndPoint + saltKey
	hash := sha256.Sum256([]byte(toHash))
	return fmt.Sprintf("%x###%s", hash, saltIndex), nil
}
