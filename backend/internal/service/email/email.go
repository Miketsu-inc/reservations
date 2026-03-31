package email

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"strings"

	"html/template"

	"github.com/BurntSushi/toml"
	"github.com/miketsu-inc/reservations/backend/emails"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/resend/resend-go/v2"
	"golang.org/x/text/language"
)

type Service struct {
	templates *template.Template
	bundle    *i18n.Bundle
	client    *resend.Client
	enabled   bool
}

func NewService(apiKey string, enableEmails bool) *Service {
	templateFS, localesFs := emails.TemplateFS()

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	mustLoadMessageFileFs(bundle, localesFs, "emails.en.toml")
	mustLoadMessageFileFs(bundle, localesFs, "emails.hu.toml")

	templates := template.New("").Funcs(template.FuncMap{
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

	return &Service{
		templates: templates,
		bundle:    bundle,
		client:    resend.NewClient(apiKey),
		enabled:   enableEmails,
	}
}

func mustLoadMessageFileFs(bundle *i18n.Bundle, fsys fs.FS, filename string) {
	data, _ := fs.ReadFile(fsys, filename)
	bundle.MustParseMessageFileBytes(data, filename)
}

func (s *Service) executeTemplate(name string, lang language.Tag, data any) string {
	var buf bytes.Buffer

	templateName := name + ".html"

	tmpl := s.templates.Lookup(templateName)
	assert.NotNil(tmpl, fmt.Sprintf("template %s not found", templateName))

	// has to be a map as passing an anonymous struct does not work
	// and passing a name struct causes you to write the name everywhere
	dataMap := utils.StructToMap(data)
	dataMap["Lang"] = lang.String()

	err := tmpl.Execute(&buf, dataMap)
	assert.Nil(err, fmt.Sprintf("error while executing template %s: %v", name, err))

	return buf.String()
}

func (s *Service) getSubject(templateName string, lang language.Tag) string {
	localizer := i18n.NewLocalizer(s.bundle, lang.String())
	return localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: fmt.Sprintf("%s.subject", templateName),
	})
}

func (s *Service) send(ctx context.Context, to string, body string, subjectText string) error {
	if !s.enabled {
		return nil
	}

	//todo: sending from our own domain, replace resend test email with address parameter of the function
	params := &resend.SendEmailRequest{
		From:    "Acme <onboarding@resend.dev>",
		To:      []string{"delivered@resend.dev"},
		Html:    body,
		Subject: subjectText,
	}

	_, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

type ForgotPasswordData struct {
	PasswordLink string `json:"password_link"`
}

func (s *Service) ForgotPassword(ctx context.Context, lang language.Tag, to string, data ForgotPasswordData) error {
	templateName := "ForgotPassword"
	subject := s.getSubject(templateName, lang)
	body := s.executeTemplate(templateName, lang, data)

	err := s.send(ctx, to, body, subject)
	if err != nil {
		return err
	}
	return nil
}

type EmailVerificationData struct {
	Code int `json:"code"`
}

func (s *Service) EmailVerification(ctx context.Context, lang language.Tag, to string, data EmailVerificationData) error {
	templateName := "EmailVerification"
	subject := s.getSubject(templateName, lang)
	body := s.executeTemplate(templateName, lang, data)

	err := s.send(ctx, to, body, subject)
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

func (s *Service) BookingConfirmation(ctx context.Context, lang language.Tag, to string, data BookingConfirmationData) error {
	templateName := "BookingConfirmation"
	subject := s.getSubject(templateName, lang)
	body := s.executeTemplate(templateName, lang, data)

	err := s.send(ctx, to, body, subject)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) BookingReminder(ctx context.Context, lang language.Tag, to string, data BookingConfirmationData) error {
	templateName := "BookingReminder"
	subject := s.getSubject(templateName, lang)
	body := s.executeTemplate(templateName, lang, data)

	err := s.send(ctx, to, body, subject)
	if err != nil {
		return err
	}

	return nil
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

func (s *Service) BookingCancellation(ctx context.Context, lang language.Tag, to string, data BookingCancellationData) error {
	templateName := "BookingCancellation"
	subject := s.getSubject(templateName, lang)
	body := s.executeTemplate(templateName, lang, data)

	err := s.send(ctx, to, body, subject)
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

func (s *Service) BookingModification(ctx context.Context, lang language.Tag, to string, data BookingModificationData) error {
	templateName := "BookingModification"
	subject := s.getSubject(templateName, lang)
	body := s.executeTemplate(templateName, lang, data)

	err := s.send(ctx, to, body, subject)
	if err != nil {
		return err
	}

	return nil
}
