package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type imageRepository struct {
	db db.DBTX
}

func NewImageRepository(db db.DBTX) domain.ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) WithTx(tx db.DBTX) domain.ImageRepository {
	return &imageRepository{db: tx}
}

func (r *imageRepository) NewImage(ctx context.Context, image domain.Image) error {
	query := `
	insert into "Image" (id, purpose, status, user_id, merchant_id, employee_id, location_id, byte_size, created_by)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query, image.Id, image.Purpose, image.Status, image.UserId, image.MerchantId, image.EmployeeId,
		image.LocationId, image.ByteSize, image.CreatedBy)
	if err != nil {
		return fmt.Errorf("NewImage: %w", err)
	}

	return nil
}

func (r *imageRepository) GetImage(ctx context.Context, imageId uuid.UUID) (domain.Image, error) {
	query := `
	select id, purpose, status, user_id, merchant_id, employee_id, location_id, byte_size, width, height, created_by, created_at
	from "Image"
	where id = $1
	`

	rows, _ := r.db.Query(ctx, query, imageId)
	image, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.Image])
	if err != nil {
		return domain.Image{}, fmt.Errorf("GetImage: %w", err)
	}

	return image, nil
}

func (r *imageRepository) UpdateImageStatus(ctx context.Context, imageId uuid.UUID, status types.ImageStatus) error {
	query := `
	update "Image" set status = $2
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, imageId, status)
	if err != nil {
		return fmt.Errorf("UpdateImageStatus: %w", err)
	}

	return nil
}

func (r *imageRepository) TransitionImageStatus(ctx context.Context, imageId uuid.UUID, from types.ImageStatus, to types.ImageStatus) (bool, error) {
	query := `
	update "Image" set status = $3
	where id = $1 and status = $2
	`

	tag, err := r.db.Exec(ctx, query, imageId, from, to)
	if err != nil {
		return false, fmt.Errorf("TransitionImageStatus: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *imageRepository) SetImageReady(ctx context.Context, imageId uuid.UUID, width int, height int) error {
	query := `
	update "Image" set status = 'ready', width = $2, height = $3
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, imageId, width, height)
	if err != nil {
		return fmt.Errorf("SetImageReady: %w", err)
	}

	return nil
}

func (r *imageRepository) DeleteImage(ctx context.Context, imageId uuid.UUID) error {
	query := `
	delete from "Image" where id = $1
	`

	_, err := r.db.Exec(ctx, query, imageId)
	if err != nil {
		return fmt.Errorf("DeleteImage: %w", err)
	}

	return nil
}

func (r *imageRepository) IsImageReferenced(ctx context.Context, imageId uuid.UUID) (bool, error) {
	query := `
	select exists (select 1 from "User" where avatar_image_id = $1)
		or exists (select 1 from "Employee" where avatar_image_id = $1)
		or exists (select 1 from "Merchant" where logo_image_id = $1)
		or exists (select 1 from "LocationImage" where image_id = $1)
	`

	var referenced bool

	err := r.db.QueryRow(ctx, query, imageId).Scan(&referenced)
	if err != nil {
		return false, fmt.Errorf("IsImageReferenced: %w", err)
	}

	return referenced, nil
}

func (r *imageRepository) GetUnusedImageIds(ctx context.Context, createdBefore time.Time, limit int) ([]uuid.UUID, error) {
	query := `
	select i.id
	from "Image" i
	where i.created_at < $1 and (
		i.status <> 'ready' or (
			not exists (select 1 from "User" where avatar_image_id = i.id)
			and not exists (select 1 from "Employee" where avatar_image_id = i.id)
			and not exists (select 1 from "Merchant" where logo_image_id = i.id)
			and not exists (select 1 from "LocationImage" where image_id = i.id)
		)
	)
	order by i.created_at
	limit $2
	`

	rows, _ := r.db.Query(ctx, query, createdBefore, limit)
	imageIds, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return nil, fmt.Errorf("GetUnusedImageIds: %w", err)
	}

	return imageIds, nil
}

func (r *imageRepository) SetUserAvatar(ctx context.Context, userId uuid.UUID, imageId *uuid.UUID) (*uuid.UUID, error) {
	query := `
	update "User" u set avatar_image_id = $2
	from (select id, avatar_image_id from "User" where id = $1 for update) old
	where u.id = old.id
	returning old.avatar_image_id
	`

	var previous *uuid.UUID

	err := r.db.QueryRow(ctx, query, userId, imageId).Scan(&previous)
	if err != nil {
		return nil, fmt.Errorf("SetUserAvatar: %w", err)
	}

	return previous, nil
}

func (r *imageRepository) SetEmployeeAvatar(ctx context.Context, merchantId uuid.UUID, employeeId int, imageId *uuid.UUID) (*uuid.UUID, error) {
	query := `
	update "Employee" e set avatar_image_id = $3
	from (select id, avatar_image_id from "Employee" where merchant_id = $1 and id = $2 for update) old
	where e.id = old.id
	returning old.avatar_image_id
	`

	var previous *uuid.UUID

	err := r.db.QueryRow(ctx, query, merchantId, employeeId, imageId).Scan(&previous)
	if err != nil {
		return nil, fmt.Errorf("SetEmployeeAvatar: %w", err)
	}

	return previous, nil
}

func (r *imageRepository) SetMerchantLogo(ctx context.Context, merchantId uuid.UUID, imageId *uuid.UUID) (*uuid.UUID, error) {
	query := `
	update "Merchant" m set logo_image_id = $2
	from (select id, logo_image_id from "Merchant" where id = $1 for update) old
	where m.id = old.id
	returning old.logo_image_id
	`

	var previous *uuid.UUID

	err := r.db.QueryRow(ctx, query, merchantId, imageId).Scan(&previous)
	if err != nil {
		return nil, fmt.Errorf("SetMerchantLogo: %w", err)
	}

	return previous, nil
}

func (r *imageRepository) NewLocationImage(ctx context.Context, locationId int, imageId uuid.UUID) error {
	query := `
	insert into "LocationImage" (image_id, location_id, sequence)
	select $2, $1, coalesce(max(sequence) + 1, 0)
	from "LocationImage"
	where location_id = $1
	`

	_, err := r.db.Exec(ctx, query, locationId, imageId)
	if err != nil {
		return fmt.Errorf("NewLocationImage: %w", err)
	}

	return nil
}

func (r *imageRepository) GetLocationImages(ctx context.Context, merchantId uuid.UUID, locationId int) ([]domain.LocationImage, error) {
	query := `
	select li.image_id, li.sequence, li.alt_text, i.width, i.height
	from "LocationImage" li
	inner join "Location" l on l.id = li.location_id
	inner join "Image" i on i.id = li.image_id
	where l.merchant_id = $1 and li.location_id = $2
	order by li.sequence, i.created_at
	`

	rows, _ := r.db.Query(ctx, query, merchantId, locationId)
	images, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.LocationImage])
	if err != nil {
		return nil, fmt.Errorf("GetLocationImages: %w", err)
	}

	return images, nil
}

func (r *imageRepository) CountLocationImages(ctx context.Context, locationId int, pendingCreatedAfter time.Time) (int, error) {
	query := `
	select
		(select count(*) from "LocationImage" where location_id = $1) +
		(select count(*) from "Image" where location_id = $1 and purpose = 'location_image'
			and (status = 'processing' or (status = 'pending' and created_at > $2)))
	`

	var count int

	err := r.db.QueryRow(ctx, query, locationId, pendingCreatedAfter).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountLocationImages: %w", err)
	}

	return count, nil
}

func (r *imageRepository) CountReadyLocationImages(ctx context.Context, locationId int) (int, error) {
	query := `
	select count(*) from "LocationImage" where location_id = $1
	`

	var count int

	err := r.db.QueryRow(ctx, query, locationId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("CountReadyLocationImages: %w", err)
	}

	return count, nil
}

func (r *imageRepository) UpdateLocationImageAltText(ctx context.Context, merchantId uuid.UUID, locationId int, imageId uuid.UUID, altText string) error {
	query := `
	update "LocationImage" li set alt_text = $4
	from "Location" l
	where l.id = li.location_id and l.merchant_id = $1 and li.location_id = $2 and li.image_id = $3
	`

	tag, err := r.db.Exec(ctx, query, merchantId, locationId, imageId, altText)
	if err != nil {
		return fmt.Errorf("UpdateLocationImageAltText: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("UpdateLocationImageAltText: %w", pgx.ErrNoRows)
	}

	return nil
}

func (r *imageRepository) DeleteLocationImage(ctx context.Context, merchantId uuid.UUID, locationId int, imageId uuid.UUID) error {
	query := `
	delete from "LocationImage" li
	using "Location" l
	where l.id = li.location_id and l.merchant_id = $1 and li.location_id = $2 and li.image_id = $3
	`

	tag, err := r.db.Exec(ctx, query, merchantId, locationId, imageId)
	if err != nil {
		return fmt.Errorf("DeleteLocationImage: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("DeleteLocationImage: %w", pgx.ErrNoRows)
	}

	return nil
}

func (r *imageRepository) ReorderLocationImages(ctx context.Context, merchantId uuid.UUID, locationId int, imageIds []uuid.UUID) error {
	query := `
	update "LocationImage" li set sequence = new_order.sequence - 1
	from unnest($3::uuid[]) with ordinality as new_order(image_id, sequence), "Location" l
	where li.image_id = new_order.image_id and l.id = li.location_id and l.merchant_id = $1 and li.location_id = $2
	`

	_, err := r.db.Exec(ctx, query, merchantId, locationId, imageIds)
	if err != nil {
		return fmt.Errorf("ReorderLocationImages: %w", err)
	}

	return nil
}
