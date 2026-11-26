package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type favoriteRepository struct {
	db db.DBTX
}

func NewFavoriteRepository(db db.DBTX) domain.FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) WithTx(tx db.DBTX) domain.FavoriteRepository {
	return &favoriteRepository{db: tx}
}

func (r *favoriteRepository) AddFavoriteLocation(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID, locationId int) error {
	query := `
	insert into "Favorite" (user_id, merchant_id, location_id) values ($1, $2, $3)
	on conflict do nothing
	`

	_, err := r.db.Exec(ctx, query, userId, merchantId, locationId)
	if err != nil {
		return fmt.Errorf("AddFavoriteLocation: %w", err)
	}

	return nil
}

func (r *favoriteRepository) RemoveFavoriteLocation(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID, locationId int) error {
	query := `
	delete from "Favorite" where user_id = $1 and merchant_id = $2 and location_id = $3
	`

	_, err := r.db.Exec(ctx, query, userId, merchantId, locationId)
	if err != nil {
		return fmt.Errorf("RemoveFavoriteLocation: %w", err)
	}

	return nil
}

func (r *favoriteRepository) GetFavoriteLocations(ctx context.Context, userId uuid.UUID) ([]domain.FavoriteLocation, error) {
	query := `
	select m.id as merchant_id, m.name as merchant_name, m.url_name as merchant_url_name,
		l.id as location_id, l.formatted_location,
		rs.average_rating, rs.review_count, img.image_id as main_image_id
	from "Favorite" f
	inner join "Merchant" m on f.merchant_id = m.id
	inner join "Location" l on f.location_id = l.id and l.is_active is true
	left join lateral (
		select coalesce(round(avg(r.rating)::numeric, 1), 0)::float8 as average_rating, count(*)::int as review_count
		from "Review" r
		where r.merchant_id = f.merchant_id and r.location_id = f.location_id and r.status <> 'removed'
	) rs on true
	left join lateral (
		select li.image_id
		from "LocationImage" li
		where li.location_id = l.id
		order by li.sequence
		limit 1
	) img on true
	where f.user_id = $1
	order by f.created_at desc
	`

	rows, _ := r.db.Query(ctx, query, userId)
	favorites, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.FavoriteLocation])
	if err != nil {
		return nil, fmt.Errorf("GetFavoriteLocations: %w", err)
	}

	return favorites, nil
}

func (r *favoriteRepository) IsFavoriteLocation(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID, locationId int) (bool, error) {
	query := `
	select exists (
		select 1 from "Favorite"
		where user_id = $1 and merchant_id = $2 and location_id = $3
	)`

	var exists bool

	err := r.db.QueryRow(ctx, query, userId, merchantId, locationId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("IsFavoriteLocation: %w", err)
	}

	return exists, nil
}
