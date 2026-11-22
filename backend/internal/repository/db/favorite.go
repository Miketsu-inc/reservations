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

func (r *favoriteRepository) AddFavoriteMerchant(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	insert into "Favorite" (user_id, merchant_id) values ($1, $2)
	on conflict do nothing
	`

	_, err := r.db.Exec(ctx, query, userId, merchantId)
	if err != nil {
		return fmt.Errorf("AddFavoriteMerchant: %w", err)
	}

	return nil
}

func (r *favoriteRepository) RemoveFavoriteMerchant(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	delete from "Favorite" where user_id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, userId, merchantId)
	if err != nil {
		return fmt.Errorf("RemoveFavoriteMerchant: %w", err)
	}

	return nil
}

func (r *favoriteRepository) GetFavoriteMerchants(ctx context.Context, userId uuid.UUID) ([]domain.FavoriteMerchant, error) {
	query := `
	select m.id as merchant_id, m.name as merchant_name, m.url_name as merchant_url_name, l.formatted_location,
		rs.average_rating, rs.review_count, img.image_id as main_image_id
	from "Favorite" f
	inner join "Merchant" m on f.merchant_id = m.id
	left join lateral (
		select id, formatted_location
		from "Location"
		where merchant_id = m.id and is_active is true
		order by sequence, id
		limit 1
	) l on true
	left join lateral (
		select coalesce(round(avg(r.rating)::numeric, 1), 0)::float8 as average_rating, count(*)::int as review_count
		from "Review" r
		where r.merchant_id = m.id and r.location_id = l.id and r.status <> 'removed'
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

	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("GetFavoriteMerchants/Query: %w", err)
	}
	favorites, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.FavoriteMerchant])
	if err != nil {
		return nil, fmt.Errorf("GetFavoriteMerchants/CollectRows: %w", err)
	}

	return favorites, nil
}

func (r *favoriteRepository) IsFavoriteMerchant(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID) (bool, error) {
	query := `
	select exists (
		select 1 from "Favorite"
		where user_id = $1 and merchant_id = $2
	)`

	var exists bool

	err := r.db.QueryRow(ctx, query, userId, merchantId).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("IsFavoriteMerchant: %w", err)
	}

	return exists, nil
}
