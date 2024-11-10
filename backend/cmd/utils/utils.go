package utils

import (
	"net/smtp"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

func SendMail(email string) error {
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	assert.True(from != "", "EMAIL_ADDRESS environment variable could not be found")
	assert.True(password != "", "EEMAIL_PASSWORD environment variable could not be found")
	assert.True(smtpPort != "", "SMTP_HOST environment variable could not be found")
	assert.True(smtpPort != "", "SMTP_PORT environment variable could not be found")

	to := []string{email}

	auth := smtp.PlainAuth("", from, password, smtpHost)

	msg := []byte("From: " + from + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: Email verification\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\r\n" +
		"\r\n" +
		"CLick on this link to verify your email address: \r\n\n" +
		"If it wasn't you who signed up with this email please ignore this message.")

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(msg))

	if err != nil {
		return err
	}
	return nil
}
