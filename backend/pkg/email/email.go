package email

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"text/template"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

const emailTemplate = `
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Email Verification</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					background-color: #f7f7f7;
					color: #333;
					padding: 20px;
				}
				.container {
					max-width: 600px;
					margin: 0 auto;
					background: white;
					border-radius: 8px;
					box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
					padding: 20px;
				}
				.btn {
					display: inline-block;
					background-color: #007BFF;
					color: white;
					padding: 10px 20px;
					border-radius: 5px;
					text-decoration: none;
					font-size: 16px;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h2>Welcome to Your Company!</h2>
				<p>We're thrilled to have you with us. To get started, please confirm your email address by clicking the button below.</p>
				<div style="text-align: center; margin: 20px 0;">
					<a href="{{.Link}}" target="_self" class="btn">Verify Your Email</a>
				</div>
				<p>If you didn’t sign up for Your Company, you can safely ignore this email.</p>
			</div>
		</body>
		</html>
	`

type verifData struct {
	Email     string
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}

var stored verifData

func genVerifToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func SendMail(email, baseURL string) error {

	token, err := genVerifToken()
	if err != nil {
		return fmt.Errorf("error generating token: %s", err.Error())
	}
	verifLink := fmt.Sprintf("%s?token=%s", baseURL, token)

	stored = verifData{
		Email:     email,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	var body bytes.Buffer
	tmpl, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %s", err.Error())
	}
	err = tmpl.Execute(&body, map[string]string{"Link": verifLink})
	if err != nil {
		return fmt.Errorf("error executing template: %s", err.Error())
	}

	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	assert.True(from != "", "EMAIL_ADDRESS environment variable could not be found")
	assert.True(password != "", "EMAIL_PASSWORD environment variable could not be found")
	assert.True(smtpPort != "", "SMTP_HOST environment variable could not be found")
	assert.True(smtpPort != "", "SMTP_PORT environment variable could not be found")

	to := []string{email}

	auth := smtp.PlainAuth("", from, password, smtpHost)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: Email Verification\n"
	msg := []byte(subject + headers + body.String())

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, []byte(msg))
	if err != nil {
		return fmt.Errorf("error sending email: %s", err)
	}

	return nil
}

func ValidateToken(newToken string) error {
	if stored.Token != newToken {
		return fmt.Errorf("the tokens don't match")
	}
	if time.Now().After(stored.ExpiresAt) {
		return fmt.Errorf("the token expired")
	}

	return nil
}
