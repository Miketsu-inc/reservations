package objectstore

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("object not found")

type Bucket interface {
	Name() string
	// Create a presigned PUT request which can be used to upload an object directly.
	// The content type and size are part of the signature, the upload fails if the client sends different values
	PresignPut(ctx context.Context, key string, contentType string, size int64, expires time.Duration) (PresignedRequest, error)
	Head(ctx context.Context, key string) (ObjectInfo, error)
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error)
	Put(ctx context.Context, input PutInput) error
	// Deleting objects that do not exist is not an error
	Delete(ctx context.Context, keys ...string) error
}

type ObjectInfo struct {
	Size        int64
	ContentType string
}

type PresignedRequest struct {
	URL    string
	Method string
	// Headers that must be sent with the request, excluding the ones set by http clients automatically (Host, Content-Length)
	Headers   http.Header
	ExpiresAt time.Time
}

type PutInput struct {
	Key          string
	Body         io.Reader
	Size         int64
	ContentType  string
	CacheControl string
}
