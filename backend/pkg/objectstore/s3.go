package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// S3 API allows deleting at most 1000 objects in a single request
const maxDeleteBatch = 1000

// Creates a client for any S3 compatible object storage
func NewS3Client(endpoint string, accessKeyId string, secretAccessKey string) *s3.Client {
	return s3.New(s3.Options{
		// required by the SDK, S3 compatible providers like R2 ignore it
		Region:       "auto",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyId, secretAccessKey, ""),
		// Without these the SDK adds checksum parameters to presigned urls which browsers can not fulfill
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
}

// Creates a client for Cloudflare R2
func NewR2Client(accountId string, accessKeyId string, secretAccessKey string) *s3.Client {
	return NewS3Client(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId), accessKeyId, secretAccessKey)
}

type S3Bucket struct {
	client  *s3.Client
	presign *s3.PresignClient
	name    string
}

func NewS3Bucket(client *s3.Client, name string) *S3Bucket {
	return &S3Bucket{
		client:  client,
		presign: s3.NewPresignClient(client),
		name:    name,
	}
}

func (b *S3Bucket) PresignPut(ctx context.Context, key string, contentType string, size int64, expires time.Duration) (PresignedRequest, error) {
	req, err := b.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(b.name),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return PresignedRequest{}, fmt.Errorf("presign put %s: %w", key, err)
	}

	headers := req.SignedHeader.Clone()
	headers.Del("Host")
	headers.Del("Content-Length")

	return PresignedRequest{
		URL:       req.URL,
		Method:    req.Method,
		Headers:   headers,
		ExpiresAt: time.Now().Add(expires),
	}, nil
}

func (b *S3Bucket) Head(ctx context.Context, key string) (ObjectInfo, error) {
	out, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return ObjectInfo{}, ErrNotFound
		}
		return ObjectInfo{}, fmt.Errorf("head %s: %w", key, err)
	}

	return ObjectInfo{
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
	}, nil
}

func (b *S3Bucket) Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	out, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, ObjectInfo{}, ErrNotFound
		}
		return nil, ObjectInfo{}, fmt.Errorf("get %s: %w", key, err)
	}

	return out.Body, ObjectInfo{
		Size:        aws.ToInt64(out.ContentLength),
		ContentType: aws.ToString(out.ContentType),
	}, nil
}

func (b *S3Bucket) Put(ctx context.Context, input PutInput) error {
	putInput := &s3.PutObjectInput{
		Bucket:        aws.String(b.name),
		Key:           aws.String(input.Key),
		Body:          input.Body,
		ContentLength: aws.Int64(input.Size),
		ContentType:   aws.String(input.ContentType),
	}

	if input.CacheControl != "" {
		putInput.CacheControl = aws.String(input.CacheControl)
	}

	_, err := b.client.PutObject(ctx, putInput)
	if err != nil {
		return fmt.Errorf("put %s: %w", input.Key, err)
	}

	return nil
}

func (b *S3Bucket) Delete(ctx context.Context, keys ...string) error {
	for start := 0; start < len(keys); start += maxDeleteBatch {
		end := min(start+maxDeleteBatch, len(keys))

		objects := make([]s3types.ObjectIdentifier, 0, end-start)
		for _, key := range keys[start:end] {
			objects = append(objects, s3types.ObjectIdentifier{Key: aws.String(key)})
		}

		out, err := b.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(b.name),
			Delete: &s3types.Delete{
				Objects: objects,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("delete objects: %w", err)
		}

		if len(out.Errors) > 0 {
			e := out.Errors[0]
			return fmt.Errorf("delete %s: %s: %s (%d failed)", aws.ToString(e.Key), aws.ToString(e.Code), aws.ToString(e.Message), len(out.Errors))
		}
	}

	return nil
}

func isNotFound(err error) bool {
	var notFound *s3types.NotFound
	var noSuchKey *s3types.NoSuchKey
	if errors.As(err, &notFound) || errors.As(err, &noSuchKey) {
		return true
	}

	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		return code == "NotFound" || code == "NoSuchKey"
	}

	var respErr interface{ HTTPStatusCode() int }
	if errors.As(err, &respErr) {
		return respErr.HTTPStatusCode() == http.StatusNotFound
	}

	return false
}
