package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

	unusedImageRetention = 24 * time.Hour
	cleanupBatchSize     = 1000
)

func (s *Service) ProcessImage(ctx context.Context, imageId uuid.UUID, crop *domain.ImageCrop, sourceContentType string) error {
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

	variants, err := s.processImage(ctx, image, crop, sourceContentType)
	if err != nil {
		if errors.Is(err, errInvalidImage) {
			return s.failImage(ctx, imageId, err)
		}
		return err
	}

	err = s.storeImage(ctx, image, variants)
	if err != nil {
		return err
	}

	// the lifecycle rule of the uploads bucket deletes it anyway if this fails
	_ = s.uploadsBucket.Delete(ctx, imageId.String())

	return nil
}

var ErrUnprocessable = errors.New("image can not be processed")

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

func (s *Service) storeImage(ctx context.Context, image domain.Image, variants []processedVariant) error {
	var width, height int

	for _, v := range variants {
		err := s.mediaBucket.Put(ctx, objectstore.PutInput{
			Key:          objectKey(image.Id, v.variant),
			Body:         bytes.NewReader(v.data),
			Size:         int64(len(v.data)),
			ContentType:  v.contentType,
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
		attached, previous, err := s.attachImage(ctx, tx, image)
		if err != nil {
			return err
		}

		if !attached {
			err = s.imageRepo.WithTx(tx).UpdateImageStatus(ctx, image.Id, types.ImageStatusFailed)
			if err != nil {
				return err
			}

			return s.scheduleImageDeletion(ctx, tx, &image.Id)
		}

		err = s.imageRepo.WithTx(tx).SetImageReady(ctx, image.Id, width, height)
		if err != nil {
			return err
		}

		if previous != nil && *previous != image.Id {
			return s.scheduleImageDeletion(ctx, tx, previous)
		}

		return nil
	})
}

// Attaches the ready image to it's target, returns false if the target no longer exists
// and returns the previous image of the target which should be deleted
func (s *Service) attachImage(ctx context.Context, tx pgx.Tx, image domain.Image) (bool, *uuid.UUID, error) {
	var previous *uuid.UUID
	var err error

	switch image.Purpose {
	case types.ImagePurposeUserAvatar:
		if image.UserId == nil {
			return false, nil, nil
		}

		previous, err = s.userRepo.WithTx(tx).SetUserAvatar(ctx, *image.UserId, &image.Id)

	case types.ImagePurposeEmployeeAvatar:
		if image.MerchantId == nil || image.EmployeeId == nil {
			return false, nil, nil
		}

		previous, err = s.teamRepo.WithTx(tx).SetEmployeeAvatar(ctx, *image.MerchantId, *image.EmployeeId, &image.Id)

	case types.ImagePurposeMerchantLogo:
		if image.MerchantId == nil {
			return false, nil, nil
		}

		previous, err = s.merchantRepo.WithTx(tx).SetMerchantLogo(ctx, *image.MerchantId, &image.Id)

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

	count, err := s.imageRepo.WithTx(tx).CountReadyLocationImages(ctx, *image.LocationId)
	if err != nil {
		return false, nil, err
	}

	// the limit could have been lowered since the upload started (e.g. subscription downgrade)
	if count >= limit {
		return false, nil, nil
	}

	err = s.imageRepo.WithTx(tx).NewLocationImage(ctx, *image.LocationId, image.Id)
	if err != nil {
		return false, nil, err
	}

	return true, nil, nil
}

func (s *Service) DeleteImage(ctx context.Context, imageId uuid.UUID) error {
	referenced, err := s.imageRepo.IsImageReferenced(ctx, imageId)
	if err != nil {
		return err
	}

	// the image was attached again (e.g. a retried job), deleting it would break it's references
	if referenced {
		return nil
	}

	err = s.mediaBucket.Delete(ctx, objectKey(imageId, VariantSmall), objectKey(imageId, VariantLarge))
	if err != nil {
		return err
	}

	err = s.uploadsBucket.Delete(ctx, imageId.String())
	if err != nil {
		return err
	}

	return s.imageRepo.DeleteImage(ctx, imageId)
}

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
