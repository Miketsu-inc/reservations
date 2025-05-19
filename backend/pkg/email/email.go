package email

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
"time"

	"html/template"

	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/emails"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/resend/resend-go/v2"
)

var templates = make(map[string]*template.Template)
var cfg *config.Config = config.LoadEnvVars()

func init() {
	templateFS := emails.TemplateFS()

err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if strings.HasSuffix(path, ".html") {
			templ, err := template.ParseFS(templateFS, path)
			assert.Nil(err, fmt.Sprintf("Failed to parse template %s: %v", path, err))

			name := strings.TrimSuffix(filepath.Base(path), ".html")
			templates[name] = templ
		}
		return nil
	})
assert.Nil(err, fmt.Sprintf("Error walking through templates: %v", err))
}

func executeTemplate(name string, data interface{}) (bytes.Buffer, error) {
	var body bytes.Buffer

	tmpl, ok := templates[name]
	if !ok {
		return body, fmt.Errorf("template %s not found", name)
	}

	err := tmpl.Execute(&body, data)
	return body, err
}

func Send(ctx context.Context, to string, body bytes.Buffer, subjectText string) error {
if !cfg.ENABLE_EMAILS {
		return nil
	}

	client := resend.NewClient(cfg.RESEND_API_TEST)

	//todo: sending from our own domain, replace resend test email with address parameter of the function
	params := &resend.SendEmailRequest{
		From:    "Acme <onboarding@resend.dev>",
		To:      []string{"delivered@resend.dev"},
		Html:    body.String(),
		Subject: subjectText,
	}

	_, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func Schedule(ctx context.Context, to string, body bytes.Buffer, subjectText string, date string) (string, error) {
	if !cfg.ENABLE_EMAILS {
return "", nil
	}

	client := resend.NewClient(cfg.RESEND_API_TEST)

	params := &resend.SendEmailRequest{
		From:        "Acme <onboarding@resend.dev>",
		To:          []string{"delivered@resend.dev"},
		Html:        body.String(),
		Subject:     subjectText,
		ScheduledAt: date,
	}

	res, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return "", err
	}

	return res.Id, nil
}

func Cancel(emailId string) error {
if !cfg.ENABLE_EMAILS {
		return nil
	}

	client := resend.NewClient(cfg.RESEND_API_TEST)

	_, err := client.Emails.Cancel(emailId)
	if err != nil {
		return err
	}
	return nil
}

func ReSchedule(emailId string, newDate string) error {
if !cfg.ENABLE_EMAILS {
		return nil
	}

	client := resend.NewClient(cfg.RESEND_API_TEST)

	updateParams := &resend.UpdateEmailRequest{
		Id:          emailId,
		ScheduledAt: newDate,
	}

	_, err := client.Emails.Update(updateParams)
	if err != nil {
		return err
	}

	return nil
}

type ForgotPasswordData struct {
	PasswordLink string `json:"password_link"`
}

func ForgotPassword(ctx context.Context, to string, data ForgotPasswordData) error {
	subject := "Állíts be új jelszót"

	body, err := executeTemplate("ForgotPassword", data)
	assert.Nil(err, fmt.Sprintf("Error executing ForgotPassword template: %s", err))

	err = Send(ctx, to, body, subject)
	if err != nil {
		return err
	}
	return nil
}

type EmailVerificationData struct {
	VerificationCode string `json:"verification_code"`
}

func EmailVerification(ctx context.Context, to string, data EmailVerificationData) error {
	subject := "Email megerősítés"

	body, err := executeTemplate("EmailVerification", data)
	assert.Nil(err, fmt.Sprintf("Error executing EmailVerification template: %s", err))

	err = Send(ctx, to, body, subject)
	if err != nil {
		return err
	}
	return nil
}

type AppointmentConfirmationData struct {
	Time        string `json:"time"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	ServiceName string `json:"service_name"`
	TimeZone    string `json:"time_zone"`
	ModifyLink  string `json:"modify_link"`
}

func AppointmentConfirmation(ctx context.Context, to string, data AppointmentConfirmationData) error {
	subject := "Időpont megerősitve"

	body, err := executeTemplate("AppointmentConfirmation", data)
	assert.Nil(err, fmt.Sprintf("Error executing AppointmentConfirmation template: %s", err))

	err = Send(ctx, to, body, subject)
	if err != nil {
		return err
	}

	return nil
}

func AppointmentReminder(ctx context.Context, to string, data AppointmentConfirmationData, date time.Time) (string, error) {
	subject := "Időpont emlékeztető"

	body, err := executeTemplate("AppointmentReminder", data)
	assert.Nil(err, fmt.Sprintf("Error executing AppointmentReminder template: %s", err))

	email_id, err := Schedule(ctx, to, body, subject, date.Format(time.RFC3339))
	if err != nil {
		return "", err
	}

	return email_id, nil
}

type AppointmentCancellationData struct {
	Time               string `json:"time"`
	Date               string `json:"date"`
	Location           string `json:"location"`
	ServiceName        string `json:"service_name"`
	TimeZone           string `json:"time_zone"`
	Reason             string `json:"reason"`
	NewAppointmentLink string `json:"new_appointment_link"`
}

func AppointmentCancellation(ctx context.Context, to string, data AppointmentCancellationData) error {
	subject := "Időpont törölve lett"

	body, err := executeTemplate("AppointmentCancellation", data)
	assert.Nil(err, fmt.Sprintf("Error executing AppointmentCancellation template: %s", err))

	err = Send(ctx, to, body, subject)
	if err != nil {
		return err
	}

	return nil
}