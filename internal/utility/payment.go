package utility

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Payload struct {
	MerchantID          string `json:"merchantId"`
	MerchantTransaction string `json:"merchantTransactionId"`
	MerchantUserID      string `json:"merchantUserId"`
	Amount              int    `json:"amount"`
	RedirectURL         string `json:"redirectUrl"`
	RedirectMode        string `json:"redirectMode"`
	CallbackURL         string `json:"callbackUrl"`
	MobileNumber        string `json:"mobileNumber"`
	PaymentInstrument   struct {
		Type string `json:"type"`
	} `json:"paymentInstrument"`
}

func PayRequest(amount int, uid string) {
	payload := Payload{
		MerchantID:          os.Getenv("MERCHANT_ID"),
		MerchantTransaction: "MT7850590068188104",
		MerchantUserID:      uid,
		Amount:              amount,
		RedirectURL:         "https://webhook.site/redirect-url",
		RedirectMode:        "REDIRECT",
		CallbackURL:         "https://webhook.site/redirect-url",
		MobileNumber:        "6376309552",
		PaymentInstrument: struct {
			Type string `json:"type"`
		}{
			Type: "PAY_PAGE",
		},
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("%s", payloadJSON)

	encodedPayload := base64.StdEncoding.EncodeToString(payloadJSON)
	saltKey := os.Getenv("MERCHANT_KEY")
	saltIndex := os.Getenv("MERCHANT_KEY_INDEX")

	// Calculate the SHA256 hash of the concatenated string
	hashInput := encodedPayload + "/pg/v1/pay" + saltKey
	// fmt.Println(hashInput)
	hash := sha256.Sum256([]byte(hashInput))
	// fmt.Printf("%x\n", hash)

	// Create the X-VERIFY header value
	xVerify := fmt.Sprintf("%v###%v", hash, saltIndex)
	// fmt.Printf("%x\n", xVerify)
	requestBody := fmt.Sprintf(`{"request":"%s"}`, encodedPayload)
	// fmt.Printf("%s", requestBody)
	buffer := bytes.NewBuffer([]byte(requestBody))

	url := "https://api-preprod.phonepe.com/apis/pg-sandbox/pg/v1/pay"
	req, _ := http.NewRequest("POST", url, buffer)

	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-VERIFY", xVerify)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(string(body))
}
