package phonepay

import (
	"os"
)

const (
	// ProdHostUrl Production Details
	ProdHostUrl = "https://api.phonepe.com/apis/hermes"

	// UatHostUrl UAT/Sandbox Details
	UatHostUrl = "https://api-preprod.phonepe.com/apis/pg-sandbox"

	// PayEndPoint End point for pay request
	PayEndPoint = "/pg/v1/pay"
)

func GetBaseEndpoint() string {
	switch os.Getenv("PHONEPAY_ENV") {
	case "production":
		return ProdHostUrl
	case "uat", "":
		return UatHostUrl
	default:
		return UatHostUrl
	}
}
