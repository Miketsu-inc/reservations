package merchantpage

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"strings"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
)

const (
	seoStartMarker = "<!--seo:start-->"
	seoEndMarker   = "<!--seo:end-->"
)

// Metadata references:
//   - Title links: https://developers.google.com/search/docs/appearance/title-link
//   - Meta descriptions: https://developers.google.com/search/docs/appearance/snippet
//   - Canonical URLs: https://developers.google.com/search/docs/crawling-indexing/consolidate-duplicate-urls
//   - Open Graph protocol: https://ogp.me/
//   - X Card markup: https://developer.x.com/en/docs/x-for-websites/cards/overview/markup
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
	baseURL   string
}

func mustNewRenderer(dist fs.FS, baseURL string) *pageRenderer {
	indexHTML, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		panic(fmt.Errorf("read tango index HTML: %w", err))
	}

	if err := validateIndexHTML(indexHTML); err != nil {
		panic(err)
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil || parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" {
		panic(fmt.Errorf("invalid tango base URL %q", baseURL))
	}
	if parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https" {
		panic(fmt.Errorf("invalid tango base URL scheme %q", parsedBaseURL.Scheme))
	}

	parsedBaseURL.RawQuery = ""
	parsedBaseURL.Fragment = ""

	return &pageRenderer{
		indexHTML: indexHTML,
		baseURL:   strings.TrimRight(parsedBaseURL.String(), "/"),
	}
}

func (r *pageRenderer) render(info domain.MerchantInfo) ([]byte, error) {
	metadata, err := newMerchantPageMetadata(info, r.baseURL)
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
