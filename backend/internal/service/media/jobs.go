package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/objectstore"
	"github.com/riverqueue/river"
)

const (
	variantCacheControl = "public, max-age=31536000, immutable"
	variantContentType  = "image/jpeg"

	// Images that never became ready or are no longer used are deleted after this
	unusedImageRetention = 24 * time.Hour
	cleanupBatchSize     = 1000
)

// attaches the ready image to it's target, returns false if the target no longer exists
// and the previous image of the target which should be deleted
type attachFunc func(ctx context.Context, tx pgx.Tx, image domain.Image) (bool, *uuid.UUID, error)

// Process an uploaded image and attach it to it's target. Returns ErrUnprocessable if the image is invalid
func (s *Service) ProcessImage(ctx context.Context, imageId uuid.UUID, crop *domain.ImageCrop) error {
	image, err := s.imageRepo.GetImage(ctx, imageId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	// already processed or failed by a previous attempt
	if image.Status != types.ImageStatusProcessing {
		return nil
	}

	data, err := s.readUpload(ctx, imageId)
	if err != nil {
		if errors.Is(err, ErrUnprocessable) {
			return s.failImage(ctx, imageId, err)
		}
		return err
	}

	variants, err := processImage(data, image.Purpose, crop)
	if err != nil {
		if errors.Is(err, errInvalidImage) {
			return s.failImage(ctx, imageId, err)
		}
		return err
	}

	err = s.storeImage(ctx, image, variants, s.attachImage)
	if err != nil {
		return err
	}

	// the lifecycle rule of the uploads bucket deletes it anyway if this fails
	_ = s.uploads.Delete(ctx, uploadKey(imageId))

	return nil
}

func (s *Service) readUpload(ctx context.Context, imageId uuid.UUID) ([]byte, error) {
	body, info, err := s.uploads.Get(ctx, uploadKey(imageId))
	if err != nil {
		if errors.Is(err, objectstore.ErrNotFound) {
			return nil, fmt.Errorf("%w: upload not found", ErrUnprocessable)
		}
		return nil, err
	}
	// nolint:errcheck
	defer body.Close()

	if info.Size > MaxUploadSize {
		return nil, fmt.Errorf("%w: upload is too large (%d bytes)", ErrUnprocessable, info.Size)
	}

	data, err := io.ReadAll(io.LimitReader(body, MaxUploadSize+1))
	if err != nil {
		return nil, fmt.Errorf("error reading upload: %w", err)
	}

	if len(data) > MaxUploadSize {
		return nil, fmt.Errorf("%w: upload is too large", ErrUnprocessable)
	}

	return data, nil
}

// Marks the image as failed, it gets deleted by the cleanup job. Returns the reason wrapped in ErrUnprocessable
func (s *Service) failImage(ctx context.Context, imageId uuid.UUID, reason error) error {
	err := s.imageRepo.UpdateImageStatus(ctx, imageId, types.ImageStatusFailed)
	if err != nil {
		return err
	}

	if errors.Is(reason, ErrUnprocessable) {
		return reason
	}

	return fmt.Errorf("%w: %w", ErrUnprocessable, reason)
}

// Upload the variants to the public bucket, then mark the image as ready and attach it to it's target
func (s *Service) storeImage(ctx context.Context, image domain.Image, variants []processedVariant, attach attachFunc) error {
	var width, height int

	for _, v := range variants {
		err := s.media.Put(ctx, objectstore.PutInput{
			Key:          objectKey(image.Id, v.variant),
			Body:         bytes.NewReader(v.data),
			Size:         int64(len(v.data)),
			ContentType:  variantContentType,
			CacheControl: variantCacheControl,
		})
		if err != nil {
			return err
		}

		if v.variant == VariantLarge {
			width, height = v.width, v.height
		}
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		imageRepo := s.imageRepo.WithTx(tx)

		attached, previous, err := attach(ctx, tx, image)
		if err != nil {
			return err
		}

		if !attached {
			err = imageRepo.UpdateImageStatus(ctx, image.Id, types.ImageStatusFailed)
			if err != nil {
				return err
			}

			return s.scheduleImageDeletion(ctx, tx, &image.Id)
		}

		err = imageRepo.SetImageReady(ctx, image.Id, width, height)
		if err != nil {
			return err
		}

		if previous != nil && *previous != image.Id {
			return s.scheduleImageDeletion(ctx, tx, previous)
		}

		return nil
	})
}

func (s *Service) attachImage(ctx context.Context, tx pgx.Tx, image domain.Image) (bool, *uuid.UUID, error) {
	imageRepo := s.imageRepo.WithTx(tx)

	var previous *uuid.UUID
	var err error

	switch image.Purpose {
	case types.ImagePurposeUserAvatar:
		if image.UserId == nil {
			return false, nil, nil
		}

		previous, err = imageRepo.SetUserAvatar(ctx, *image.UserId, &image.Id)

	case types.ImagePurposeEmployeeAvatar:
		if image.MerchantId == nil || image.EmployeeId == nil {
			return false, nil, nil
		}

		previous, err = imageRepo.SetEmployeeAvatar(ctx, *image.MerchantId, *image.EmployeeId, &image.Id)

	case types.ImagePurposeMerchantLogo:
		if image.MerchantId == nil {
			return false, nil, nil
		}

		previous, err = imageRepo.SetMerchantLogo(ctx, *image.MerchantId, &image.Id)

	case types.ImagePurposeLocationImage:
		return s.attachLocationImage(ctx, tx, image)

	default:
		return false, nil, fmt.Errorf("unknown image purpose: %s", image.Purpose)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, err
	}

	return true, previous, nil
}

func (s *Service) attachLocationImage(ctx context.Context, tx pgx.Tx, image domain.Image) (bool, *uuid.UUID, error) {
	if image.MerchantId == nil || image.LocationId == nil {
		return false, nil, nil
	}

	// serializes attaching images to the same location, so concurrent jobs can not exceed the limit
	// or get the same sequence
	err := s.locationRepo.WithTx(tx).LockLocation(ctx, *image.LocationId, *image.MerchantId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, err
	}

	limit, err := s.locationImageLimit(ctx, *image.MerchantId)
	if err != nil {
		return false, nil, err
	}

	imageRepo := s.imageRepo.WithTx(tx)

	count, err := imageRepo.CountReadyLocationImages(ctx, *image.LocationId)
	if err != nil {
		return false, nil, err
	}

	// the limit could have been lowered since the upload started (e.g. subscription downgrade)
	if count >= limit {
		return false, nil, nil
	}

	err = imageRepo.NewLocationImage(ctx, *image.LocationId, image.Id)
	if err != nil {
		return false, nil, err
	}

	return true, nil, nil
}

// Delete the image's objects and database row, if it is not referenced anymore
func (s *Service) DeleteImage(ctx context.Context, imageId uuid.UUID) error {
	referenced, err := s.imageRepo.IsImageReferenced(ctx, imageId)
	if err != nil {
		return err
	}

	// the image was attached again (e.g. a retried job), deleting it would break it's references
	if referenced {
		return nil
	}

	err = s.media.Delete(ctx, objectKey(imageId, VariantSmall), objectKey(imageId, VariantLarge))
	if err != nil {
		return err
	}

	err = s.uploads.Delete(ctx, uploadKey(imageId))
	if err != nil {
		return err
	}

	return s.imageRepo.DeleteImage(ctx, imageId)
}

// Schedule the deletion of images which never became ready (abandoned uploads, failed processing) or are no longer used
// (e.g. the user, employee, location or merchant was deleted)
func (s *Service) CleanupImages(ctx context.Context) error {
	imageIds, err := s.imageRepo.GetUnusedImageIds(ctx, time.Now().UTC().Add(-unusedImageRetention), cleanupBatchSize)
	if err != nil {
		return err
	}

	if len(imageIds) == 0 {
		return nil
	}

	params := make([]river.InsertManyParams, len(imageIds))
	for i, id := range imageIds {
		params[i] = river.InsertManyParams{Args: args.DeleteImage{ImageId: id}}
	}

	_, err = s.enqueuer.InsertManyFast(ctx, params)
	if err != nil {
		return fmt.Errorf("error scheduling image deletions: %w", err)
	}

	return nil
}
