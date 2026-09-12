package merchants

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
	indexDocument []byte
	publicBaseURL string
}

func newPageRenderer(dist fs.FS, publicBaseURL string) (*pageRenderer, error) {
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

	return &pageRenderer{
		indexDocument: indexDocument,
		publicBaseURL: strings.TrimRight(parsedBaseURL.String(), "/"),
	}, nil
}

func (r *pageRenderer) render(info domain.MerchantInfo) ([]byte, error) {
	metadata, err := newMerchantPageMetadata(info, r.publicBaseURL)
	if err != nil {
		return nil, err
	}

	var head bytes.Buffer
	if err := merchantHeadTemplate.Execute(&head, metadata); err != nil {
		return nil, fmt.Errorf("render merchant metadata: %w", err)
	}

	return replaceSEOHead(r.indexDocument, head.Bytes()), nil
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
