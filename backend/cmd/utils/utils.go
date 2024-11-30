package utils

import (
	"net/smtp"

	_ "github.com/joho/godotenv/autoload"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
)

func SendMail(email string) error {
	cfg := config.LoadEnvVars()

	from := cfg.EMAIL_ADDRESS
	password := cfg.EMAIL_PASSWORD
	smtpHost := cfg.SMTP_HOST
	smtpPort := cfg.SMTP_PORT

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
