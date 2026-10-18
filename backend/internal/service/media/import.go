package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

// Hosts the oauth providers serve profile pictures from. Downloading from any other host is refused
// to prevent the server from being used to make requests to arbitrary (e.g. internal) addresses
var allowedAvatarHostSuffixes = []string{
	".googleusercontent.com",
	".fbcdn.net",
	".fbsbx.com",
}

func isAllowedAvatarURL(u *url.URL) bool {
	if u.Scheme != "https" || u.Port() != "" {
		return false
	}

	host := strings.ToLower(u.Hostname())
	for _, suffix := range allowedAvatarHostSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}

	return false
}

func newAvatarHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}

			if !isAllowedAvatarURL(req.URL) {
				return fmt.Errorf("redirect to disallowed host: %s", req.URL.Host)
			}

			return nil
		},
	}
}

// Import the profile picture of a user who signed up with an oauth provider, unless they already have one.
// Returns ErrUnprocessable if the picture can not be imported
func (s *Service) ImportOauthAvatar(ctx context.Context, userId uuid.UUID, rawURL string) error {
	data, err := s.downloadAvatar(ctx, rawURL)
	if err != nil {
		return err
	}

	imageId, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("error creating image id: %w", err)
	}

	image := domain.Image{
		Id:        imageId,
		Purpose:   types.ImagePurposeUserAvatar,
		Status:    types.ImageStatusProcessing,
		UserId:    &userId,
		ByteSize:  int64(len(data)),
		CreatedBy: &userId,
	}

	err = s.imageRepo.NewImage(ctx, image)
	if err != nil {
		return err
	}

	variants, err := processImage(data, image.Purpose, nil)
	if err != nil {
		if errors.Is(err, errInvalidImage) {
			return s.failImage(ctx, imageId, err)
		}
		return err
	}

	return s.storeImage(ctx, image, variants, func(ctx context.Context, tx pgx.Tx, image domain.Image) (bool, *uuid.UUID, error) {
		// the user could have uploaded a picture in the meantime
		attached, err := s.imageRepo.WithTx(tx).SetUserAvatarIfEmpty(ctx, userId, image.Id)
		return attached, nil, err
	})
}

func (s *Service) downloadAvatar(ctx context.Context, rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil || !isAllowedAvatarURL(u) {
		return nil, fmt.Errorf("%w: disallowed avatar url", ErrUnprocessable)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating avatar request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error downloading avatar: %w", err)
	}
	// nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest && resp.StatusCode < http.StatusInternalServerError {
		return nil, fmt.Errorf("%w: avatar download returned status %s", ErrUnprocessable, resp.Status)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("avatar download returned status %s", resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxUploadSize+1))
	if err != nil {
		return nil, fmt.Errorf("error reading avatar: %w", err)
	}

	if len(data) > MaxUploadSize {
		return nil, fmt.Errorf("%w: avatar is too large", ErrUnprocessable)
	}

	return data, nil
}
