package media

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/keys"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/objectstore"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
	"github.com/redis/go-redis/v9"
)

const (
	uploadURLExpiry = 10 * time.Minute

	uploadRateLimit  = 30
	uploadRateWindow = time.Hour
)

// Maximum number of images per location
var locationImageLimits = map[types.SubTier]int{
	types.SubTierFree:       3,
	types.SubTierPro:        15,
	types.SubTierEnterprise: 30,
}

type Service struct {
	imageRepo    domain.ImageRepository
	teamRepo     domain.TeamRepository
	locationRepo domain.LocationRepository
	merchantRepo domain.MerchantRepository
	// private bucket, the clients upload the original files here
	uploads objectstore.Bucket
	// public bucket containing the processed variants
	media      objectstore.Bucket
	kv         *redis.Client
	httpClient *http.Client
	enqueuer   queue.Enqueuer
	txManager  db.TransactionManager
}

func NewService(image domain.ImageRepository, team domain.TeamRepository, location domain.LocationRepository, merchant domain.MerchantRepository,
	uploads objectstore.Bucket, media objectstore.Bucket, kv *redis.Client, enqueuer queue.Enqueuer, txManager db.TransactionManager) *Service {
	return &Service{
		imageRepo:    image,
		teamRepo:     team,
		locationRepo: location,
		merchantRepo: merchant,
		uploads:      uploads,
		media:        media,
		kv:           kv,
		httpClient:   newAvatarHTTPClient(),
		enqueuer:     enqueuer,
		txManager:    txManager,
	}
}

func (s *Service) SetEnqueuer(client queue.Enqueuer) {
	s.enqueuer = client
}

type UploadInput struct {
	ContentType types.ImageContentType
	Size        int64
}

type Upload struct {
	ImageId uuid.UUID
	Request objectstore.PresignedRequest
}

func (s *Service) CreateUserAvatarUpload(ctx context.Context, input UploadInput) (Upload, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	return s.createUpload(ctx, domain.Image{
		Purpose:   types.ImagePurposeUserAvatar,
		UserId:    &userId,
		CreatedBy: &userId,
	}, input)
}

func (s *Service) CreateEmployeeAvatarUpload(ctx context.Context, employeeId int, input UploadInput) (Upload, error) {
	actor := actor.MustGetFromContext(ctx)

	if err := s.checkEmployeeAvatarAccess(ctx, actor, employeeId); err != nil {
		return Upload{}, err
	}

	return s.createUpload(ctx, domain.Image{
		Purpose:    types.ImagePurposeEmployeeAvatar,
		MerchantId: &actor.MerchantId,
		EmployeeId: &employeeId,
		CreatedBy:  &actor.UserId,
	}, input)
}

func (s *Service) CreateMerchantLogoUpload(ctx context.Context, input UploadInput) (Upload, error) {
	actor := actor.MustGetFromContext(ctx)

	return s.createUpload(ctx, domain.Image{
		Purpose:    types.ImagePurposeMerchantLogo,
		MerchantId: &actor.MerchantId,
		CreatedBy:  &actor.UserId,
	}, input)
}

func (s *Service) CreateLocationImageUpload(ctx context.Context, locationId int, input UploadInput) (Upload, error) {
	actor := actor.MustGetFromContext(ctx)

	_, err := s.locationRepo.GetLocation(ctx, locationId, actor.MerchantId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Upload{}, ErrLocationNotFound
		}
		return Upload{}, err
	}

	limit, err := s.locationImageLimit(ctx, actor.MerchantId)
	if err != nil {
		return Upload{}, err
	}

	// pending uploads which can not be completed anymore are not counted
	count, err := s.imageRepo.CountLocationImages(ctx, locationId, time.Now().UTC().Add(-uploadURLExpiry))
	if err != nil {
		return Upload{}, err
	}

	if count >= limit {
		return Upload{}, apperr.Wrap(ErrLocationImageLimitReached, nil).With("limit", limit)
	}

	return s.createUpload(ctx, domain.Image{
		Purpose:    types.ImagePurposeLocationImage,
		MerchantId: &actor.MerchantId,
		LocationId: &locationId,
		CreatedBy:  &actor.UserId,
	}, input)
}

func (s *Service) createUpload(ctx context.Context, image domain.Image, input UploadInput) (Upload, error) {
	if input.Size <= 0 || input.Size > MaxUploadSize {
		return Upload{}, ErrInvalidUploadSize
	}

	if err := s.checkUploadRateLimit(ctx, *image.CreatedBy); err != nil {
		return Upload{}, err
	}

	imageId, err := uuid.NewV7()
	if err != nil {
		return Upload{}, fmt.Errorf("error creating image id: %w", err)
	}

	image.Id = imageId
	image.Status = types.ImageStatusPending
	image.ByteSize = input.Size

	err = s.imageRepo.NewImage(ctx, image)
	if err != nil {
		return Upload{}, err
	}

	req, err := s.uploads.PresignPut(ctx, uploadKey(imageId), input.ContentType.String(), input.Size, uploadURLExpiry)
	if err != nil {
		return Upload{}, err
	}

	return Upload{
		ImageId: imageId,
		Request: req,
	}, nil
}

// Fixed window rate limit on the number of uploads a user can start
func (s *Service) checkUploadRateLimit(ctx context.Context, userId uuid.UUID) error {
	key := keys.ImageUploadRate{UserId: userId}.String()

	pipe := s.kv.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.ExpireNX(ctx, key, uploadRateWindow)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("error incrementing upload rate limit: %w", err)
	}

	if incr.Val() > uploadRateLimit {
		return ErrUploadRateLimited
	}

	return nil
}

// Owners and admins can change anyone's profile picture, staff members only their own
func (s *Service) checkEmployeeAvatarAccess(ctx context.Context, actor actor.EmployeeContext, employeeId int) error {
	if actor.Role == types.EmployeeRoleStaff && actor.EmployeeId != employeeId {
		return ErrEmployeeAvatarForbidden
	}

	_, err := s.teamRepo.GetEmployee(ctx, actor.MerchantId, employeeId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEmployeeNotFound
		}
		return err
	}

	return nil
}

func (s *Service) locationImageLimit(ctx context.Context, merchantId uuid.UUID) (int, error) {
	tier, err := s.merchantRepo.GetMerchantSubscriptionTier(ctx, merchantId)
	if err != nil {
		return 0, err
	}

	return locationImageLimits[tier], nil
}

// Returns the image if it was created by the current user
func (s *Service) getOwnImage(ctx context.Context, imageId uuid.UUID) (domain.Image, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	image, err := s.imageRepo.GetImage(ctx, imageId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Image{}, ErrImageNotFound
		}
		return domain.Image{}, err
	}

	if image.CreatedBy == nil || *image.CreatedBy != userId {
		return domain.Image{}, ErrImageNotFound
	}

	return image, nil
}

// Called by the client after the file was uploaded with the presigned request, starts the processing of the image
func (s *Service) CompleteUpload(ctx context.Context, imageId uuid.UUID, crop *domain.ImageCrop) error {
	image, err := s.getOwnImage(ctx, imageId)
	if err != nil {
		return err
	}

	if image.Status != types.ImageStatusPending {
		return ErrUploadAlreadyCompleted
	}

	info, err := s.uploads.Head(ctx, uploadKey(imageId))
	if err != nil {
		if errors.Is(err, objectstore.ErrNotFound) {
			return ErrUploadMissing
		}
		return err
	}

	_, contentTypeErr := types.NewImageContentType(info.ContentType)
	if info.Size != image.ByteSize || info.Size > MaxUploadSize || contentTypeErr != nil {
		err = s.imageRepo.UpdateImageStatus(ctx, imageId, types.ImageStatusFailed)
		if err != nil {
			return err
		}

		return ErrUploadInvalid
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		started, err := s.imageRepo.WithTx(tx).TransitionImageStatus(ctx, imageId, types.ImageStatusPending, types.ImageStatusProcessing)
		if err != nil {
			return err
		}

		if !started {
			return ErrUploadAlreadyCompleted
		}

		_, err = s.enqueuer.InsertTx(ctx, tx, args.ProcessImage{
			ImageId: imageId,
			Crop:    crop,
		}, nil)
		if err != nil {
			return fmt.Errorf("error scheduling image processing: %w", err)
		}

		return nil
	})
}

// Get an image created by the current user, used to poll the processing status
func (s *Service) GetUpload(ctx context.Context, imageId uuid.UUID) (domain.Image, error) {
	return s.getOwnImage(ctx, imageId)
}

func (s *Service) RemoveUserAvatar(ctx context.Context) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		previous, err := s.imageRepo.WithTx(tx).SetUserAvatar(ctx, userId, nil)
		if err != nil {
			return err
		}

		return s.scheduleImageDeletion(ctx, tx, previous)
	})
}

func (s *Service) RemoveEmployeeAvatar(ctx context.Context, employeeId int) error {
	actor := actor.MustGetFromContext(ctx)

	if err := s.checkEmployeeAvatarAccess(ctx, actor, employeeId); err != nil {
		return err
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		previous, err := s.imageRepo.WithTx(tx).SetEmployeeAvatar(ctx, actor.MerchantId, employeeId, nil)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrEmployeeNotFound
			}
			return err
		}

		return s.scheduleImageDeletion(ctx, tx, previous)
	})
}

func (s *Service) RemoveMerchantLogo(ctx context.Context) error {
	actor := actor.MustGetFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		previous, err := s.imageRepo.WithTx(tx).SetMerchantLogo(ctx, actor.MerchantId, nil)
		if err != nil {
			return err
		}

		return s.scheduleImageDeletion(ctx, tx, previous)
	})
}

type LocationImages struct {
	Images []domain.LocationImage
	Limit  int
}

func (s *Service) GetLocationImages(ctx context.Context, locationId int) (LocationImages, error) {
	actor := actor.MustGetFromContext(ctx)

	_, err := s.locationRepo.GetLocation(ctx, locationId, actor.MerchantId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LocationImages{}, ErrLocationNotFound
		}
		return LocationImages{}, err
	}

	images, err := s.imageRepo.GetLocationImages(ctx, actor.MerchantId, locationId)
	if err != nil {
		return LocationImages{}, err
	}

	limit, err := s.locationImageLimit(ctx, actor.MerchantId)
	if err != nil {
		return LocationImages{}, err
	}

	return LocationImages{
		Images: images,
		Limit:  limit,
	}, nil
}

func (s *Service) UpdateLocationImage(ctx context.Context, locationId int, imageId uuid.UUID, altText string) error {
	actor := actor.MustGetFromContext(ctx)

	err := s.imageRepo.UpdateLocationImageAltText(ctx, actor.MerchantId, locationId, imageId, altText)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLocationImageNotFound
		}
		return err
	}

	return nil
}

func (s *Service) DeleteLocationImage(ctx context.Context, locationId int, imageId uuid.UUID) error {
	actor := actor.MustGetFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := s.imageRepo.WithTx(tx).DeleteLocationImage(ctx, actor.MerchantId, locationId, imageId)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrLocationImageNotFound
			}
			return err
		}

		return s.scheduleImageDeletion(ctx, tx, &imageId)
	})
}

func (s *Service) ReorderLocationImages(ctx context.Context, locationId int, imageIds []uuid.UUID) error {
	actor := actor.MustGetFromContext(ctx)

	return s.imageRepo.ReorderLocationImages(ctx, actor.MerchantId, locationId, imageIds)
}

func (s *Service) scheduleImageDeletion(ctx context.Context, tx pgx.Tx, imageId *uuid.UUID) error {
	if imageId == nil {
		return nil
	}

	_, err := s.enqueuer.InsertTx(ctx, tx, args.DeleteImage{ImageId: *imageId}, nil)
	if err != nil {
		return fmt.Errorf("error scheduling image deletion: %w", err)
	}

	return nil
}
