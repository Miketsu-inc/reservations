package imgproxy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

var ErrInvalidSource = errors.New("invalid source image")

// Client of an imgproxy server, see https://docs.imgproxy.net
type Client struct {
	baseURL    string
	key        []byte
	salt       []byte
	httpClient *http.Client
}

// The key and salt are hex encoded, they are used to sign the urls
func NewClient(baseURL string, key string, salt string) *Client {
	keyBytes, err := hex.DecodeString(key)
	assert.Nil(err, "imgproxy key is not hex encoded", err)

	saltBytes, err := hex.DecodeString(salt)
	assert.Nil(err, "imgproxy salt is not hex encoded", err)

	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		key:        keyBytes,
		salt:       saltBytes,
		httpClient: &http.Client{Timeout: time.Minute},
	}
}

// The source is an url imgproxy can read from, e.g. s3://bucket/key
func (c *Client) Process(ctx context.Context, source string, options []string, format string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url(source, options, format), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating imgproxy request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling imgproxy: %w", err)
	}
	// nolint:errcheck
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading imgproxy response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	// the source could not be found or decoded, or it is too large
	case http.StatusNotFound, http.StatusUnprocessableEntity:
		return nil, fmt.Errorf("%w: %s", ErrInvalidSource, body)
	default:
		return nil, fmt.Errorf("imgproxy returned status %s: %s", resp.Status, body)
	}
}

func (c *Client) url(source string, options []string, format string) string {
	encodedSource := base64.RawURLEncoding.EncodeToString([]byte(source))
	path := fmt.Sprintf("/%s/%s.%s", strings.Join(options, "/"), encodedSource, format)

	mac := hmac.New(sha256.New, c.key)
	mac.Write(c.salt)
	mac.Write([]byte(path))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return c.baseURL + "/" + signature + path
}
