package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type ImageRepository interface {
	WithTx(tx db.DBTX) ImageRepository

	NewImage(ctx context.Context, image Image) error
	GetImage(ctx context.Context, imageId uuid.UUID) (Image, error)
	UpdateImageStatus(ctx context.Context, imageId uuid.UUID, status types.ImageStatus) error
	// Update the image status only if it is currently in the from status. Returns false if the status was not updated
	TransitionImageStatus(ctx context.Context, imageId uuid.UUID, from types.ImageStatus, to types.ImageStatus) (bool, error)
	// Set the image status to ready together with the dimensions of the largest variant
	SetImageReady(ctx context.Context, imageId uuid.UUID, width int, height int) error
	DeleteImage(ctx context.Context, imageId uuid.UUID) error

	// Wether the image is used as an avatar, logo or location image
	IsImageReferenced(ctx context.Context, imageId uuid.UUID) (bool, error)
	// Get images created before the given time that never became ready or are no longer referenced
	GetUnusedImageIds(ctx context.Context, createdBefore time.Time, limit int) ([]uuid.UUID, error)

	// Set the User's avatar, returns the previous avatar. Returns pgx.ErrNoRows if the User does not exist
	SetUserAvatar(ctx context.Context, userId uuid.UUID, imageId *uuid.UUID) (*uuid.UUID, error)
	// Set the User's avatar only if they do not have one yet. Returns false if the avatar was not set
	SetUserAvatarIfEmpty(ctx context.Context, userId uuid.UUID, imageId uuid.UUID) (bool, error)
	// Set the Employee's own avatar, returns the previous avatar. Returns pgx.ErrNoRows if the Employee does not exist
	SetEmployeeAvatar(ctx context.Context, merchantId uuid.UUID, employeeId int, imageId *uuid.UUID) (*uuid.UUID, error)
	// Set the Merchant's logo, returns the previous logo. Returns pgx.ErrNoRows if the Merchant does not exist
	SetMerchantLogo(ctx context.Context, merchantId uuid.UUID, imageId *uuid.UUID) (*uuid.UUID, error)

	// Insert the image as the last image of the location
	NewLocationImage(ctx context.Context, locationId int, imageId uuid.UUID) error
	GetLocationImages(ctx context.Context, merchantId uuid.UUID, locationId int) ([]LocationImage, error)
	// Number of images of the location, including the ones which are uploading or being processed
	// (pending images only count if they were created after pendingCreatedAfter)
	CountLocationImages(ctx context.Context, locationId int, pendingCreatedAfter time.Time) (int, error)
	// Number of ready images of the location
	CountReadyLocationImages(ctx context.Context, locationId int) (int, error)
	UpdateLocationImageAltText(ctx context.Context, merchantId uuid.UUID, locationId int, imageId uuid.UUID, altText string) error
	// Returns pgx.ErrNoRows if the image does not belong to the location
	DeleteLocationImage(ctx context.Context, merchantId uuid.UUID, locationId int, imageId uuid.UUID) error
	// Set the sequence of the location images in the order of the given ids
	ReorderLocationImages(ctx context.Context, merchantId uuid.UUID, locationId int, imageIds []uuid.UUID) error
}

type Image struct {
	Id         uuid.UUID          `db:"id"`
	Purpose    types.ImagePurpose `db:"purpose"`
	Status     types.ImageStatus  `db:"status"`
	UserId     *uuid.UUID         `db:"user_id"`
	MerchantId *uuid.UUID         `db:"merchant_id"`
	EmployeeId *int               `db:"employee_id"`
	LocationId *int               `db:"location_id"`
	ByteSize   int64              `db:"byte_size"`
	Width      *int               `db:"width"`
	Height     *int               `db:"height"`
	CreatedBy  *uuid.UUID         `db:"created_by"`
	CreatedAt  time.Time          `db:"created_at"`
}

// Crop rectangle in pixels, relative to the image as it is displayed (after applying it's EXIF orientation)
type ImageCrop struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type LocationImage struct {
	ImageId  uuid.UUID `json:"image_id" db:"image_id"`
	Sequence int       `json:"sequence" db:"sequence"`
	AltText  string    `json:"alt_text" db:"alt_text"`
	Width    int       `json:"width" db:"width"`
	Height   int       `json:"height" db:"height"`
}
