package utility

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateRandomCode() string {
	// Set the seed value for random number generation using the current time
	rand.Seed(time.Now().UnixNano())

	// Generate a random number between 100000 and 999999 (inclusive)
	randomNumber := rand.Intn(900000) + 100000

	// Convert the random number to a string with leading zeros
	randomCode := fmt.Sprintf("%06d", randomNumber)

	return randomCode
}
