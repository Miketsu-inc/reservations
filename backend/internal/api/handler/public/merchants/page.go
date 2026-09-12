package merchants

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
)

const (
	seoStartMarker = "<!--seo:start-->"
	seoEndMarker   = "<!--seo:end-->"
)

var seoTemplate = template.Must(template.New("seo").Parse(`<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}" />
<link rel="canonical" href="{{.CanonicalURL}}" />
<meta property="og:type" content="website" />
<meta property="og:site_name" content="Reservations" />
<meta property="og:title" content="{{.Title}}" />
<meta property="og:description" content="{{.Description}}" />
<meta property="og:url" content="{{.CanonicalURL}}" />
<meta name="twitter:card" content="summary" />
<meta name="twitter:title" content="{{.Title}}" />
<meta name="twitter:description" content="{{.Description}}" />
<script type="application/ld+json">{{.StructuredData}}</script>`))

var notFoundTemplate = template.Must(template.New("not-found").Parse(`<title>Merchant not found | Reservations</title>
<meta name="robots" content="noindex, nofollow" />`))

type merchantInfoService interface {
	GetInfo(ctx context.Context, merchantName string) (domain.MerchantInfo, error)
}

type PageHandler struct {
	service       merchantInfoService
	indexDocument []byte
	publicBaseURL string
}

type seoData struct {
	Title          string
	Description    string
	CanonicalURL   template.URL
	StructuredData template.JS
}

type localBusinessSchema struct {
	Context      string                      `json:"@context"`
	Type         string                      `json:"@type"`
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	URL          string                      `json:"url"`
	Email        string                      `json:"email,omitempty"`
	Address      *postalAddressSchema        `json:"address,omitempty"`
	Geo          *geoCoordinatesSchema       `json:"geo,omitempty"`
	OpeningHours []openingHoursSpecification `json:"openingHoursSpecification,omitempty"`
}

type postalAddressSchema struct {
	Type            string `json:"@type"`
	StreetAddress   string `json:"streetAddress,omitempty"`
	AddressLocality string `json:"addressLocality,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	AddressCountry  string `json:"addressCountry,omitempty"`
}

type geoCoordinatesSchema struct {
	Type      string  `json:"@type"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type openingHoursSpecification struct {
	Type      string `json:"@type"`
	DayOfWeek string `json:"dayOfWeek"`
	Opens     string `json:"opens"`
	Closes    string `json:"closes"`
}

func NewPageHandler(service merchantInfoService, dist fs.FS, publicBaseURL string) (*PageHandler, error) {
	indexDocument, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		return nil, fmt.Errorf("read Tango index document: %w", err)
	}

	if err := validateIndexDocument(indexDocument); err != nil {
		return nil, err
	}

	parsedBaseURL, err := url.Parse(publicBaseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" {
		return nil, fmt.Errorf("invalid Tango public URL %q", publicBaseURL)
	}
	if parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid Tango public URL scheme %q", parsedBaseURL.Scheme)
	}
	parsedBaseURL.RawQuery = ""
	parsedBaseURL.Fragment = ""

	return &PageHandler{
		service:       service,
		indexDocument: indexDocument,
		publicBaseURL: strings.TrimRight(parsedBaseURL.String(), "/"),
	}, nil
}

func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	merchantName := chi.URLParam(r, "merchantName")
	if merchantName == "" {
		h.writeNotFound(w)
		return
	}

	info, err := h.service.GetInfo(r.Context(), merchantName)
	if err != nil {
		if errors.Is(err, merchantServ.ErrMerchantNotFound) {
			h.writeNotFound(w)
			return
		}

		slog.ErrorContext(r.Context(), "render merchant page", "merchant", merchantName, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	document, err := h.renderMerchantPage(info)
	if err != nil {
		slog.ErrorContext(r.Context(), "render merchant page metadata", "merchant", merchantName, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writeHTML(w, http.StatusOK, document)
}

func (h *PageHandler) renderMerchantPage(info domain.MerchantInfo) ([]byte, error) {
	canonicalURL, err := url.JoinPath(h.publicBaseURL, "m", info.UrlName)
	if err != nil {
		return nil, fmt.Errorf("build canonical URL: %w", err)
	}

	description := merchantDescription(info)
	structuredData, err := json.Marshal(buildLocalBusinessSchema(info, canonicalURL, description))
	if err != nil {
		return nil, fmt.Errorf("marshal merchant structured data: %w", err)
	}

	var head bytes.Buffer
	err = seoTemplate.Execute(&head, seoData{
		Title:          fmt.Sprintf("%s | Book online | Reservations", info.Name),
		Description:    description,
		CanonicalURL:   template.URL(canonicalURL),
		StructuredData: template.JS(structuredData),
	})
	if err != nil {
		return nil, fmt.Errorf("render merchant metadata: %w", err)
	}

	return replaceSEOHead(h.indexDocument, head.Bytes()), nil
}

func (h *PageHandler) writeNotFound(w http.ResponseWriter) {
	var head bytes.Buffer
	if err := notFoundTemplate.Execute(&head, nil); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	writeHTML(w, http.StatusNotFound, replaceSEOHead(h.indexDocument, head.Bytes()))
}

func validateIndexDocument(document []byte) error {
	start := bytes.Index(document, []byte(seoStartMarker))
	end := bytes.Index(document, []byte(seoEndMarker))
	if start == -1 || end == -1 || end <= start {
		return fmt.Errorf("tango index document does not contain a valid SEO marker block")
	}
	return nil
}

func replaceSEOHead(document, head []byte) []byte {
	start := bytes.Index(document, []byte(seoStartMarker))
	end := bytes.Index(document, []byte(seoEndMarker)) + len(seoEndMarker)

	result := make([]byte, 0, len(document)+len(head)-(end-start))
	result = append(result, document[:start]...)
	result = append(result, head...)
	result = append(result, document[end:]...)
	return result
}

func merchantDescription(info domain.MerchantInfo) string {
	description := normalizeWhitespace(info.Introduction)
	if description == "" {
		description = normalizeWhitespace(info.AboutUs)
	}
	if description == "" {
		description = fmt.Sprintf("Book an appointment with %s at %s.", info.Name, info.FormattedLocation)
	}

	return truncateUTF8(description, 160)
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func truncateUTF8(value string, maxRunes int) string {
	if utf8.RuneCountInString(value) <= maxRunes {
		return value
	}

	runes := []rune(value)
	return strings.TrimSpace(string(runes[:maxRunes-1])) + "…"
}

func buildLocalBusinessSchema(info domain.MerchantInfo, canonicalURL, description string) localBusinessSchema {
	schema := localBusinessSchema{
		Context:     "https://schema.org",
		Type:        "LocalBusiness",
		Name:        info.Name,
		Description: description,
		URL:         canonicalURL,
		Email:       info.ContactEmail,
	}

	address := postalAddressSchema{Type: "PostalAddress"}
	if info.Address != nil {
		address.StreetAddress = *info.Address
	}
	if info.City != nil {
		address.AddressLocality = *info.City
	}
	if info.PostalCode != nil {
		address.PostalCode = *info.PostalCode
	}
	if info.Country != nil {
		address.AddressCountry = *info.Country
	}
	if address.StreetAddress != "" || address.AddressLocality != "" || address.PostalCode != "" || address.AddressCountry != "" {
		schema.Address = &address
	}

	if info.GeoPoint.Lat != 0 || info.GeoPoint.Lon != 0 {
		schema.Geo = &geoCoordinatesSchema{
			Type:      "GeoCoordinates",
			Latitude:  info.GeoPoint.Lat,
			Longitude: info.GeoPoint.Lon,
		}
	}

	dayNames := [...]string{
		"https://schema.org/Sunday",
		"https://schema.org/Monday",
		"https://schema.org/Tuesday",
		"https://schema.org/Wednesday",
		"https://schema.org/Thursday",
		"https://schema.org/Friday",
		"https://schema.org/Saturday",
	}
	for day, slots := range info.BusinessHours {
		if day < 0 || day >= len(dayNames) {
			continue
		}
		for _, slot := range slots {
			schema.OpeningHours = append(schema.OpeningHours, openingHoursSpecification{
				Type:      "OpeningHoursSpecification",
				DayOfWeek: dayNames[day],
				Opens:     slot.StartTime.Format("15:04"),
				Closes:    slot.EndTime.Format("15:04"),
			})
		}
	}

	return schema
}

func writeHTML(w http.ResponseWriter, status int, document []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(document)
}
