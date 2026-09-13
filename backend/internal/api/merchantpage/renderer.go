package merchantpage

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"strings"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

const (
	seoStartMarker = "<!--seo:start-->"
	seoEndMarker   = "<!--seo:end-->"
)

var merchantHeadTemplate = template.Must(template.New("merchant-head").Parse(`<title>{{.Title}}</title>
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

type pageRenderer struct {
	indexHTML []byte
	baseUrl   string
}

func newRenderer(dist fs.FS, baseUrl string) *pageRenderer {
	indexHTML, err := fs.ReadFile(dist, "index.html") // will this fail?
	assert.Nil(err, "failed to read tango index html")

	err = validateIndexHTML(indexHTML)
	assert.Nil(err, "error validating index html")

	parsedBaseURL, err := url.Parse(baseUrl)
	assert.Nil(err, "invalid tango base url", baseUrl)
	assert.True(parsedBaseURL.Scheme != "", "invalid tango base url", baseUrl)
	assert.True(parsedBaseURL.Host != "", "invalid tango base url", baseUrl)
	assert.True(parsedBaseURL.Scheme == "http" || parsedBaseURL.Scheme == "https", "invalid tango base url scheme", baseUrl)

	parsedBaseURL.RawQuery = ""
	parsedBaseURL.Fragment = ""

	return &pageRenderer{
		indexHTML: indexHTML,
		baseUrl:   strings.TrimRight(parsedBaseURL.String(), "/"),
	}
}

func (r *pageRenderer) render(info domain.MerchantInfo) ([]byte, error) {
	metadata, err := newMerchantPageMetadata(info, r.baseUrl)
	if err != nil {
		return nil, err
	}

	var head bytes.Buffer
	if err := merchantHeadTemplate.Execute(&head, metadata); err != nil {
		return nil, fmt.Errorf("render merchant metadata: %w", err)
	}

	return replaceSEOHead(r.indexHTML, head.Bytes()), nil
}

func validateIndexHTML(html []byte) error {
	start := bytes.Index(html, []byte(seoStartMarker))
	end := bytes.Index(html, []byte(seoEndMarker))
	if start == -1 || end == -1 || end <= start {
		return fmt.Errorf("tango index html does not contain a valid SEO marker block")
	}
	return nil
}

func replaceSEOHead(html, head []byte) []byte {
	start := bytes.Index(html, []byte(seoStartMarker))
	end := bytes.Index(html, []byte(seoEndMarker)) + len(seoEndMarker)

	result := make([]byte, 0, len(html)+len(head)-(end-start))
	result = append(result, html[:start]...)
	result = append(result, head...)
	result = append(result, html[end:]...)
	return result
}
