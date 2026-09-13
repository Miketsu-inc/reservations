package merchantpage

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
)

type pageMetadata struct {
	Title          string
	Description    string
	CanonicalURL   template.URL
	StructuredData template.JS
}

// Structured data references:
//   - Google's supported LocalBusiness fields and eligibility requirements:
//     https://developers.google.com/search/docs/appearance/structured-data/local-business
//   - Full vocabulary: https://schema.org/LocalBusiness
//   - Postal addresses: https://schema.org/PostalAddress
//   - Geographic coordinates: https://schema.org/GeoCoordinates
//   - Opening hours: https://schema.org/OpeningHoursSpecification
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

func newMerchantPageMetadata(info domain.MerchantInfo, baseURL string) (pageMetadata, error) {
	merchantPageURL, err := url.JoinPath(baseURL, "m", info.UrlName)
	if err != nil {
		return pageMetadata{}, fmt.Errorf("build merchant page URL: %w", err)
	}

	description := formatMerchantDescription(info)

	structuredData, err := json.Marshal(buildLocalBusinessSchema(info, merchantPageURL, description))
	if err != nil {
		return pageMetadata{}, fmt.Errorf("marshal merchant structured data: %w", err)
	}

	return pageMetadata{
		Title:        fmt.Sprintf("%s | Book online | Reservations", info.Name),
		Description:  description,
		CanonicalURL: template.URL(merchantPageURL),
		// json.Marshal escapes HTML-significant characters before this value is
		// embedded as JSON in a script element.
		StructuredData: template.JS(structuredData),
	}, nil
}

func formatMerchantDescription(info domain.MerchantInfo) string {
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

	return utils.TruncateUTF8(description, 160)
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func buildLocalBusinessSchema(info domain.MerchantInfo, merchantPageURL, description string) localBusinessSchema {
	schema := localBusinessSchema{
		Context:      "https://schema.org",
		Type:         "LocalBusiness",
		Name:         info.Name,
		Description:  description,
		URL:          merchantPageURL,
		Email:        info.ContactEmail,
		Address:      buildAddress(info),
		Geo:          buildGeo(info),
		OpeningHours: buildOpeningHours(info),
	}

	return schema
}

func buildAddress(info domain.MerchantInfo) *postalAddressSchema {
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
		return &address
	}

	return nil
}

func buildGeo(info domain.MerchantInfo) *geoCoordinatesSchema {
	if info.GeoPoint.Lat != 0 || info.GeoPoint.Lon != 0 {
		return &geoCoordinatesSchema{
			Type:      "GeoCoordinates",
			Latitude:  info.GeoPoint.Lat,
			Longitude: info.GeoPoint.Lon,
		}
	}

	return nil
}

func buildOpeningHours(info domain.MerchantInfo) []openingHoursSpecification {
	var openingHours []openingHoursSpecification

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
			openingHours = append(openingHours, openingHoursSpecification{
				Type:      "OpeningHoursSpecification",
				DayOfWeek: dayNames[day],
				Opens:     slot.StartTime.Format("15:04"),
				Closes:    slot.EndTime.Format("15:04"),
			})
		}
	}

	return openingHours
}
