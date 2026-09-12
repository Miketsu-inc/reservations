package merchants

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

type pageMetadata struct {
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

func newMerchantPageMetadata(info domain.MerchantInfo, publicBaseURL string) (pageMetadata, error) {
	canonicalURL, err := url.JoinPath(publicBaseURL, "m", info.UrlName)
	if err != nil {
		return pageMetadata{}, fmt.Errorf("build canonical URL: %w", err)
	}

	description := merchantDescription(info)
	structuredData, err := json.Marshal(buildLocalBusinessSchema(info, canonicalURL, description))
	if err != nil {
		return pageMetadata{}, fmt.Errorf("marshal merchant structured data: %w", err)
	}

	return pageMetadata{
		Title:        fmt.Sprintf("%s | Book online | Reservations", info.Name),
		Description:  description,
		CanonicalURL: template.URL(canonicalURL),
		// json.Marshal escapes HTML-significant characters before this value is
		// embedded as JSON in a script element.
		StructuredData: template.JS(structuredData),
	}, nil
}

func merchantDescription(info domain.MerchantInfo) string {
	description := normalizeWhitespace(info.Introduction)
	if description == "" {
		description = normalizeWhitespace(info.AboutUs)
	}
	if description == "" && info.FormattedLocation != "" {
		description = fmt.Sprintf("Book an appointment with %s at %s.", info.Name, info.FormattedLocation)
	}
	if description == "" {
		description = fmt.Sprintf("Book an appointment with %s online.", info.Name)
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
