package email

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"html/template"

	"github.com/BurntSushi/toml"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
	"github.com/miketsu-inc/reservations/backend/emails"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/resend/resend-go/v2"
	"golang.org/x/text/language"
)

var templates *template.Template
var bundle *i18n.Bundle

func init() {
	templateFS, localesFs := emails.TemplateFS()

	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	mustLoadMessageFileFs(localesFs, "emails.en.toml")
	mustLoadMessageFileFs(localesFs, "emails.hu.toml")

	templates = template.New("").Funcs(template.FuncMap{
		"T": func(lang, key string, data ...any) string {
			var templateData any
			if len(data) > 0 {
				templateData = data[0]
			}

			localizer := i18n.NewLocalizer(bundle, lang)
			msg := localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID:    key,
				TemplateData: templateData,
			})
			return msg
		},
	})
	err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".html") {
			_, err := templates.ParseFS(templateFS, path)
			assert.Nil(err, fmt.Sprintf("Failed to parse template %s: %v", path, err))
		}
		return nil
	})
	assert.Nil(err, fmt.Sprintf("Error walking through templates: %v", err))
}

func mustLoadMessageFileFs(fsys fs.FS, filename string) {
	data, _ := fs.ReadFile(fsys, filename)
	bundle.MustParseMessageFileBytes(data, filename)
}

func executeTemplate(name string, lang language.Tag, data any) string {
	var buf bytes.Buffer

	templateName := name + ".html"

	tmpl := templates.Lookup(templateName)
	assert.NotNil(tmpl, fmt.Sprintf("template %s not found", templateName))

	// has to be a map as passing an anonymous struct does not work
	// and passing a name struct causes you to write the name everywhere
	dataMap := utils.StructToMap(data)
	dataMap["Lang"] = lang.String()

	err := tmpl.Execute(&buf, dataMap)
	assert.Nil(err, fmt.Sprintf("error while executing template %s: %v", name, err))

	return buf.String()
}

func getSubject(templateName string, lang language.Tag) string {
	localizer := i18n.NewLocalizer(bundle, lang.String())
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: fmt.Sprintf("%s.subject", templateName),
	})
}

func send(ctx context.Context, to string, body string, subjectText string) error {
	if !config.LoadEnvVars().ENABLE_EMAILS {
		return nil
	}

	client := resend.NewClient(config.LoadEnvVars().RESEND_API_TEST)

	//todo: sending from our own domain, replace resend test email with address parameter of the function
	params := &resend.SendEmailRequest{
		From:    "Acme <onboarding@resend.dev>",
		To:      []string{"delivered@resend.dev"},
		Html:    body,
		Subject: subjectText,
	}

	_, err := client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func schedule(ctx context.Context, to string, body string, subjectText string, date string) (string, error) {
	if !config.LoadEnvVars().ENABLE_EMAILS {
		return "", nil
	}

	client := resend.NewClient(config.LoadEnvVars().RESEND_API_TEST)

	params := &resend.SendEmailRequest{
		From:        "Acme <onboarding@resend.dev>",
		To:          []string{"delivered@resend.dev"},
		Html:        body,
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
	if !config.LoadEnvVars().ENABLE_EMAILS {
		return nil
	}

	client := resend.NewClient(config.LoadEnvVars().RESEND_API_TEST)

	_, err := client.Emails.Cancel(emailId)
	if err != nil {
		return err
	}
	return nil
}

func ReSchedule(emailId string, newDate string) error {
	if !config.LoadEnvVars().ENABLE_EMAILS {
		return nil
	}

	client := resend.NewClient(config.LoadEnvVars().RESEND_API_TEST)

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

func ForgotPassword(ctx context.Context, lang language.Tag, to string, data ForgotPasswordData) error {
	templateName := "ForgotPassword"
	subject := getSubject(templateName, lang)
	body := executeTemplate(templateName, lang, data)

	err := send(ctx, to, body, subject)
	if err != nil {
		return err
	}
	return nil
}

type EmailVerificationData struct {
	Code int `json:"code"`
}

func EmailVerification(ctx context.Context, lang language.Tag, to string, data EmailVerificationData) error {
	templateName := "EmailVerification"
	subject := getSubject(templateName, lang)
	body := executeTemplate(templateName, lang, data)

	err := send(ctx, to, body, subject)
	if err != nil {
		return err
	}
	return nil
}

type BookingConfirmationData struct {
	Time        string `json:"time"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	ServiceName string `json:"service_name"`
	TimeZone    string `json:"time_zone"`
	ModifyLink  string `json:"modify_link"`
}

func BookingConfirmation(ctx context.Context, lang language.Tag, to string, data BookingConfirmationData) error {
	templateName := "BookingConfirmation"
	subject := getSubject(templateName, lang)
	body := executeTemplate(templateName, lang, data)

	err := send(ctx, to, body, subject)
	if err != nil {
		return err
	}

	return nil
}

func BookingReminder(ctx context.Context, lang language.Tag, to string, data BookingConfirmationData, date time.Time) (string, error) {
	templateName := "BookingReminder"
	subject := getSubject(templateName, lang)
	body := executeTemplate(templateName, lang, data)

	email_id, err := schedule(ctx, to, body, subject, date.Format(time.RFC3339))
	if err != nil {
		return "", err
	}

	return email_id, nil
}

type BookingCancellationData struct {
	Time           string `json:"time"`
	Date           string `json:"date"`
	Location       string `json:"location"`
	ServiceName    string `json:"service_name"`
	TimeZone       string `json:"time_zone"`
	Reason         string `json:"reason"`
	NewBookingLink string `json:"new_booking_link"`
}

func BookingCancellation(ctx context.Context, lang language.Tag, to string, data BookingCancellationData) error {
	templateName := "BookingCancellation"
	subject := getSubject(templateName, lang)
	body := executeTemplate(templateName, lang, data)

	err := send(ctx, to, body, subject)
	if err != nil {
		return err
	}

	return nil
}

type BookingModificationData struct {
	Time        string `json:"time"`
	Date        string `json:"date"`
	Location    string `json:"location"`
	ServiceName string `json:"service_name"`
	TimeZone    string `json:"time_zone"`
	ModifyLink  string `json:"modify_link"`
	OldTime     string `json:"old_time"`
	OldDate     string `json:"old_date"`
}

func BookingModification(ctx context.Context, lang language.Tag, to string, data BookingModificationData) error {
	templateName := "BookingModification"
	subject := getSubject(templateName, lang)
	body := executeTemplate(templateName, lang, data)

	err := send(ctx, to, body, subject)
	if err != nil {
		return err
	}

	return nil
}
