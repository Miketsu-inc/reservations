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
	TransitionImageStatus(ctx context.Context, imageId uuid.UUID, from types.ImageStatus, to types.ImageStatus) (bool, error)
	SetImageReady(ctx context.Context, imageId uuid.UUID, width int, height int) error
	DeleteImage(ctx context.Context, imageId uuid.UUID) error

	IsImageReferenced(ctx context.Context, imageId uuid.UUID) (bool, error)
	GetUnusedImageIds(ctx context.Context, createdBefore time.Time, limit int) ([]uuid.UUID, error)

	SetUserAvatar(ctx context.Context, userId uuid.UUID, imageId *uuid.UUID) (*uuid.UUID, error)
	SetEmployeeAvatar(ctx context.Context, merchantId uuid.UUID, employeeId int, imageId *uuid.UUID) (*uuid.UUID, error)
	SetMerchantLogo(ctx context.Context, merchantId uuid.UUID, imageId *uuid.UUID) (*uuid.UUID, error)

	NewLocationImage(ctx context.Context, locationId int, imageId uuid.UUID) error
	GetLocationImages(ctx context.Context, merchantId uuid.UUID, locationId int) ([]LocationImage, error)
	CountLocationImages(ctx context.Context, locationId int, pendingCreatedAfter time.Time) (int, error)
	CountReadyLocationImages(ctx context.Context, locationId int) (int, error)
	UpdateLocationImageAltText(ctx context.Context, merchantId uuid.UUID, locationId int, imageId uuid.UUID, altText string) error
	DeleteLocationImage(ctx context.Context, merchantId uuid.UUID, locationId int, imageId uuid.UUID) error
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
