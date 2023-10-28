package utility

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendMail(msg string, receiver string, subject string) error {
	from := os.Getenv("MAILING_ADDRESS")
	password := os.Getenv("MAILING_SERVICE_PSWD")

	// SMTP server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	auth := smtp.PlainAuth("", from, password, smtpHost)

	return smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{receiver}, buildMessage(from, receiver, subject, msg))
}

func buildMessage(from string, to string, subject string, msg string) []byte {
	return []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", from, to, subject, msg))
}
