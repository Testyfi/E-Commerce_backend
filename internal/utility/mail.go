package utility

import (
	"net/smtp"
	"os"
)

func SendMail(msg string, receiver string, subject string) error {
	// Sender data.
	from := os.Getenv("MAILING_ADDRESS")
	password := os.Getenv("MAILING_SERVICE_PSWD")

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	//  Message.
	message := []byte(
		"From: Testify\r\n" +
			"To: " + receiver + "\r\n" +
			"Subject: " + subject + "\r\n\r\n" +
			msg,
	)

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// // Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{receiver}, message)
	if err != nil {
		return err
	}

	return nil
}
