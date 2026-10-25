package media

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/miketsu-inc/reservations/backend/pkg/imgproxy"
	"github.com/miketsu-inc/reservations/backend/pkg/objectstore"
	"github.com/miketsu-inc/reservations/backend/pkg/queue"
)

const (
	uploadURLExpiry = 10 * time.Minute
)

var locationImageLimits = map[types.SubTier]int{
	types.SubTierFree:       3,
	types.SubTierPro:        15,
	types.SubTierEnterprise: 30,
}

type Service struct {
	imageRepo     domain.ImageRepository
	userRepo      domain.UserRepository
	teamRepo      domain.TeamRepository
	locationRepo  domain.LocationRepository
	merchantRepo  domain.MerchantRepository
	uploadsBucket objectstore.Bucket
	mediaBucket   objectstore.Bucket
	imgproxy      *imgproxy.Client
	enqueuer      queue.Enqueuer
	txManager     db.TransactionManager
}

func NewService(image domain.ImageRepository, user domain.UserRepository, team domain.TeamRepository, location domain.LocationRepository, merchant domain.MerchantRepository,
	uploadsBucket objectstore.Bucket, mediaBucket objectstore.Bucket, imgproxy *imgproxy.Client, enqueuer queue.Enqueuer, txManager db.TransactionManager) *Service {
	return &Service{
		imageRepo:     image,
		userRepo:      user,
		teamRepo:      team,
		locationRepo:  location,
		merchantRepo:  merchant,
		uploadsBucket: uploadsBucket,
		mediaBucket:   mediaBucket,
		imgproxy:      imgproxy,
		enqueuer:      enqueuer,
		txManager:     txManager,
	}
}

func (s *Service) SetEnqueuer(client queue.Enqueuer) {
	s.enqueuer = client
}

type UploadInput struct {
	ContentType string
	Size        int64
}

type UploadResult struct {
	ImageId uuid.UUID
	Request objectstore.PresignedRequest
}

func (s *Service) CreateUserAvatarUpload(ctx context.Context, input UploadInput) (UploadResult, error) {
	userId := jwt.MustGetUserIDFromContext(ctx)

	return s.createUpload(ctx, domain.Image{
		Purpose:   types.ImagePurposeUserAvatar,
		UserId:    &userId,
		CreatedBy: &userId,
	}, input)
}

func (s *Service) CreateEmployeeAvatarUpload(ctx context.Context, employeeId int, input UploadInput) (UploadResult, error) {
	actor := actor.MustGetFromContext(ctx)

	if err := s.checkEmployeeAvatarAccess(ctx, actor, employeeId); err != nil {
		return UploadResult{}, err
	}

	return s.createUpload(ctx, domain.Image{
		Purpose:    types.ImagePurposeEmployeeAvatar,
		MerchantId: &actor.MerchantId,
		EmployeeId: &employeeId,
		CreatedBy:  &actor.UserId,
	}, input)
}

func (s *Service) CreateMerchantLogoUpload(ctx context.Context, input UploadInput) (UploadResult, error) {
	actor := actor.MustGetFromContext(ctx)

	return s.createUpload(ctx, domain.Image{
		Purpose:    types.ImagePurposeMerchantLogo,
		MerchantId: &actor.MerchantId,
		CreatedBy:  &actor.UserId,
	}, input)
}

func (s *Service) CreateLocationImageUpload(ctx context.Context, locationId int, input UploadInput) (UploadResult, error) {
	actor := actor.MustGetFromContext(ctx)

	_, err := s.locationRepo.GetLocation(ctx, locationId, actor.MerchantId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UploadResult{}, ErrLocationNotFound
		}
		return UploadResult{}, err
	}

	limit, err := s.locationImageLimit(ctx, actor.MerchantId)
	if err != nil {
		return UploadResult{}, err
	}

	// pending uploads which can not be completed anymore are not counted
	count, err := s.imageRepo.CountLocationImages(ctx, locationId, time.Now().UTC().Add(-uploadURLExpiry))
	if err != nil {
		return UploadResult{}, err
	}

	if count >= limit {
		return UploadResult{}, apperr.Wrap(ErrLocationImageLimitReached, nil).With("limit", limit)
	}

	return s.createUpload(ctx, domain.Image{
		Purpose:    types.ImagePurposeLocationImage,
		MerchantId: &actor.MerchantId,
		LocationId: &locationId,
		CreatedBy:  &actor.UserId,
	}, input)
}

func (s *Service) createUpload(ctx context.Context, image domain.Image, input UploadInput) (UploadResult, error) {
	if input.Size <= 0 || input.Size > MaxUploadSize {
		return UploadResult{}, ErrInvalidUploadSize
	}

	imageId, err := uuid.NewV7()
	if err != nil {
		return UploadResult{}, fmt.Errorf("error creating image id: %w", err)
	}

	image.Id = imageId
	image.Status = types.ImageStatusPending
	image.ByteSize = input.Size

	err = s.imageRepo.NewImage(ctx, image)
	if err != nil {
		return UploadResult{}, err
	}

	req, err := s.uploadsBucket.PresignPut(ctx, imageId.String(), input.ContentType, input.Size, uploadURLExpiry)
	if err != nil {
		return UploadResult{}, err
	}

	return UploadResult{
		ImageId: imageId,
		Request: req,
	}, nil
}

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

// the browser calls this after it uploaded the image through the presigned url
func (s *Service) CompleteUpload(ctx context.Context, imageId uuid.UUID, crop *domain.ImageCrop) error {
	image, err := s.getOwnImage(ctx, imageId)
	if err != nil {
		return err
	}

	if image.Status != types.ImageStatusPending {
		return ErrUploadAlreadyCompleted
	}

	info, err := s.uploadsBucket.Head(ctx, imageId.String())
	if err != nil {
		if errors.Is(err, objectstore.ErrNotFound) {
			return ErrUploadMissing
		}
		return err
	}

	isAllowedType := slices.Contains(allowedContentTypes, info.ContentType)
	if info.Size != image.ByteSize || info.Size > MaxUploadSize || !isAllowedType {
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
			ImageId:     imageId,
			Crop:        crop,
			ContentType: info.ContentType,
		}, nil)
		if err != nil {
			return fmt.Errorf("error scheduling image processing: %w", err)
		}

		return nil
	})
}

func (s *Service) GetUpload(ctx context.Context, imageId uuid.UUID) (domain.Image, error) {
	return s.getOwnImage(ctx, imageId)
}

func (s *Service) RemoveUserAvatar(ctx context.Context) error {
	userId := jwt.MustGetUserIDFromContext(ctx)

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		previous, err := s.userRepo.WithTx(tx).SetUserAvatar(ctx, userId, nil)
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
		previous, err := s.teamRepo.WithTx(tx).SetEmployeeAvatar(ctx, actor.MerchantId, employeeId, nil)
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
		previous, err := s.merchantRepo.WithTx(tx).SetMerchantLogo(ctx, actor.MerchantId, nil)
		if err != nil {
			return err
		}

		return s.scheduleImageDeletion(ctx, tx, previous)
	})
}

type GetLocationImagesResult struct {
	Images []domain.LocationImage
	Limit  int
}

func (s *Service) GetLocationImages(ctx context.Context, locationId int) (GetLocationImagesResult, error) {
	actor := actor.MustGetFromContext(ctx)

	_, err := s.locationRepo.GetLocation(ctx, locationId, actor.MerchantId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return GetLocationImagesResult{}, ErrLocationNotFound
		}
		return GetLocationImagesResult{}, err
	}

	images, err := s.imageRepo.GetLocationImages(ctx, actor.MerchantId, locationId)
	if err != nil {
		return GetLocationImagesResult{}, err
	}

	limit, err := s.locationImageLimit(ctx, actor.MerchantId)
	if err != nil {
		return GetLocationImagesResult{}, err
	}

	return GetLocationImagesResult{
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

	_, err := s.locationRepo.GetLocation(ctx, locationId, actor.MerchantId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrLocationNotFound
		}
		return err
	}

	images, err := s.imageRepo.GetLocationImages(ctx, actor.MerchantId, locationId)
	if err != nil {
		return err
	}

	if !isSameImageSet(images, imageIds) {
		return ErrInvalidImageOrder
	}

	return s.imageRepo.ReorderLocationImages(ctx, actor.MerchantId, locationId, imageIds)
}

func isSameImageSet(images []domain.LocationImage, imageIds []uuid.UUID) bool {
	if len(images) != len(imageIds) {
		return false
	}

	remaining := make(map[uuid.UUID]bool, len(images))
	for _, img := range images {
		remaining[img.ImageId] = true
	}

	for _, id := range imageIds {
		if !remaining[id] {
			return false
		}
		delete(remaining, id)
	}

	return true
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
