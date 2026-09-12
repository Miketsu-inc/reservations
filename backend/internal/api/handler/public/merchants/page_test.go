package merchants

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testIndexDocument = `<!doctype html>
<html><head><!--seo:start--><title>Reservations</title><!--seo:end--></head>
<body><div id="root"></div><script type="module" src="/assets/app.js"></script></body></html>`

type stubMerchantInfoService struct {
	info          domain.MerchantInfo
	err           error
	requestedName string
}

func (s *stubMerchantInfoService) GetInfo(_ context.Context, merchantName string) (domain.MerchantInfo, error) {
	s.requestedName = merchantName
	return s.info, s.err
}

func TestPageHandlerRendersMerchantMetadata(t *testing.T) {
	t.Parallel()

	street := "1 Main Street"
	city := "Budapest"
	postalCode := "1011"
	country := "HU"
	service := &stubMerchantInfoService{info: domain.MerchantInfo{
		Name:              `M&M </title><script>alert("x")</script>`,
		UrlName:           "m-and-m",
		ContactEmail:      "hello@example.com",
		Introduction:      "Book   a relaxing\n appointment.",
		Address:           &street,
		City:              &city,
		PostalCode:        &postalCode,
		Country:           &country,
		FormattedLocation: "1 Main Street, Budapest",
		GeoPoint:          types.GeoPoint{Lat: 47.4979, Lon: 19.0402},
		BusinessHours: domain.BusinessHours{
			1: {{StartTime: businessHour("09:00"), EndTime: businessHour("17:30")}},
		},
	}}
	handler := newTestPageHandler(t, service, "https://reservations.example/base/")

	response := requestMerchantPage(handler, "requested-slug")
	body := response.Body.String()

	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "text/html; charset=utf-8", response.Header().Get("Content-Type"))
	assert.Equal(t, "requested-slug", service.requestedName)
	assert.Contains(t, body, `<title>M&amp;M &lt;/title&gt;&lt;script&gt;alert(&#34;x&#34;)&lt;/script&gt; | Book online | Reservations</title>`)
	assert.Contains(t, body, `<meta name="description" content="Book a relaxing appointment." />`)
	assert.Contains(t, body, `<link rel="canonical" href="https://reservations.example/base/m/m-and-m" />`)
	assert.Contains(t, body, `<meta property="og:title"`)
	assert.Contains(t, body, `<meta name="twitter:card" content="summary" />`)
	assert.NotContains(t, body, `<script>alert("x")</script>`)
	assert.Equal(t, 1, strings.Count(body, "<title>"))
	assert.Contains(t, body, `<script type="module" src="/assets/app.js"></script>`)

	structuredData := extractStructuredData(t, body)
	assert.Equal(t, "https://schema.org", structuredData["@context"])
	assert.Equal(t, "LocalBusiness", structuredData["@type"])
	assert.Equal(t, service.info.Name, structuredData["name"])
	assert.Equal(t, "https://reservations.example/base/m/m-and-m", structuredData["url"])
	assert.Equal(t, "hello@example.com", structuredData["email"])

	address, ok := structuredData["address"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, street, address["streetAddress"])
	assert.Equal(t, country, address["addressCountry"])

	hours, ok := structuredData["openingHoursSpecification"].([]any)
	require.True(t, ok)
	require.Len(t, hours, 1)
	monday, ok := hours[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "https://schema.org/Monday", monday["dayOfWeek"])
	assert.Equal(t, "09:00", monday["opens"])
	assert.Equal(t, "17:30", monday["closes"])
}

func TestPageHandlerReturnsSEOAwareNotFoundPage(t *testing.T) {
	t.Parallel()

	service := &stubMerchantInfoService{
		err: apperr.Wrap(merchantServ.ErrMerchantNotFound, errors.New("missing row")),
	}
	handler := newTestPageHandler(t, service, "https://reservations.example")

	response := requestMerchantPage(handler, "missing")
	body := response.Body.String()

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Contains(t, body, `<title>Merchant not found | Reservations</title>`)
	assert.Contains(t, body, `<meta name="robots" content="noindex, nofollow" />`)
	assert.Contains(t, body, `<script type="module" src="/assets/app.js"></script>`)
	assert.NotContains(t, body, seoStartMarker)
}

func TestNewPageHandlerValidatesInputs(t *testing.T) {
	t.Parallel()

	service := &stubMerchantInfoService{}

	_, err := NewPageHandler(service, fstest.MapFS{
		"index.html": {Data: []byte("<html><head><title>Reservations</title></head></html>")},
	}, "https://reservations.example")
	assert.ErrorContains(t, err, "SEO marker block")

	_, err = NewPageHandler(service, testIndexFS(), "javascript:alert(1)")
	assert.ErrorContains(t, err, "invalid Tango public URL")
}

func TestMerchantDescriptionFallbackAndTruncation(t *testing.T) {
	t.Parallel()

	info := domain.MerchantInfo{
		Name:              "Merchant",
		FormattedLocation: "Budapest",
	}
	assert.Equal(t, "Book an appointment with Merchant at Budapest.", merchantDescription(info))

	info.Introduction = strings.Repeat("é", 200)
	description := merchantDescription(info)
	assert.Equal(t, 160, len([]rune(description)))
	assert.True(t, strings.HasSuffix(description, "…"))
}

func newTestPageHandler(t *testing.T, service merchantInfoService, publicBaseURL string) *PageHandler {
	t.Helper()
	handler, err := NewPageHandler(service, testIndexFS(), publicBaseURL)
	require.NoError(t, err)
	return handler
}

func testIndexFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html": {Data: []byte(testIndexDocument)},
	}
}

func requestMerchantPage(handler http.Handler, merchantName string) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	router.Get("/m/{merchantName}", handler.ServeHTTP)

	request := httptest.NewRequest(http.MethodGet, "/m/"+merchantName, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func extractStructuredData(t *testing.T, document string) map[string]any {
	t.Helper()
	const startTag = `<script type="application/ld+json">`
	start := strings.Index(document, startTag)
	require.NotEqual(t, -1, start)
	start += len(startTag)
	end := strings.Index(document[start:], "</script>")
	require.NotEqual(t, -1, end)

	var data map[string]any
	require.NoError(t, json.Unmarshal([]byte(document[start:start+end]), &data))
	return data
}

func businessHour(value string) time.Time {
	parsed, _ := time.Parse("15:04", value)
	return parsed
}
