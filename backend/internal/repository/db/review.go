package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type reviewRepository struct {
	db db.DBTX
}

func NewReviewRepository(db db.DBTX) domain.ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) WithTx(tx db.DBTX) domain.ReviewRepository {
	return &reviewRepository{db: tx}
}

func (r *reviewRepository) NewReview(ctx context.Context, review domain.Review) error {
	query := `
	insert into "Review" (booking_id, user_id, merchant_id, location_id, employee_id, service_id, service_name, rating, comment) 
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query, review.BookingId, review.UserId, review.MerchantId, review.LocationId, review.EmployeeId, review.ServiceId, review.ServiceName,
		review.Rating, review.Comment)
	if err != nil {
		return fmt.Errorf("CreateReview: %w", err)
	}

	return nil
}

func (r *reviewRepository) UpdateReview(ctx context.Context, reviewId int, userId uuid.UUID, bookingId *int, rating int16, comment *string) error {
	query := `
	update "Review" 
	set rating = $1, comment = $2, booking_id = coalesce($3, booking_id), updated_at = now()
	where id = $4 and user_id = $5 and status = 'published'
	`

	_, err := r.db.Exec(ctx, query, rating, comment, bookingId, reviewId, userId)
	if err != nil {
		return fmt.Errorf("UpdateReview: %w", err)
	}

	return nil
}

func (r *reviewRepository) DeleteReview(ctx context.Context, reviewId int, userId uuid.UUID) error {
	query := `
	delete from "Review" where id = $1 and user_id = $2 and status = 'published'
	`

	_, err := r.db.Exec(ctx, query, reviewId, userId)
	if err != nil {
		return fmt.Errorf("DeleteReview: %w", err)
	}

	return nil
}

func (r *reviewRepository) ReportReview(ctx context.Context, reviewId int, merchantId uuid.UUID, reportReason string) error {
	query := `
	update "Review" set status = 'reported', report_reason = $1, updated_at = now()
	where id = $2 and merchant_id = $3
	`

	_, err := r.db.Exec(ctx, query, reportReason, reviewId, merchantId)
	if err != nil {
		return fmt.Errorf("ReportReview: %w", err)
	}

	return nil
}

func (r *reviewRepository) RemoveReview(ctx context.Context, reviewId int, moderatedAt time.Time) error {
	query := `
	update "Review" set status = 'removed', moderated_at = $1, updated_at = now()
	where id = $2
	`

	_, err := r.db.Exec(ctx, query, moderatedAt, reviewId)
	if err != nil {
		return fmt.Errorf("RemoveReview: %w", err)
	}

	return nil
}

func (r *reviewRepository) GetReviewById(ctx context.Context, reviewId int) (domain.Review, error) {
	query := `
	select id, booking_id, user_id, merchant_id, location_id, employee_id, service_id, service_name, rating, comment, status, report_reason, moderated_at, created_at, updated_at
	from "Review"
	where id = $1
	`

	rows, _ := r.db.Query(ctx, query, reviewId)
	review, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.Review])
	if err != nil {
		return domain.Review{}, fmt.Errorf("GetReviewById: %w", err)
	}

	return review, nil
}

func (r *reviewRepository) GetExistingReviewCombination(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID, locationId int, employeeId int, serviceId int) (*domain.Review, error) {
	query := `
	select id, booking_id, user_id, merchant_id, location_id, employee_id, service_id, service_name, rating, comment, status, report_reason, moderated_at, created_at, updated_at
	from "Review"
	where user_id = $1 and merchant_id = $2 and location_id = $3 and employee_id = $4 and service_id = $5
	order by created_at desc
	limit 1
	`

	rows, _ := r.db.Query(ctx, query, userId, merchantId, locationId, employeeId, serviceId)
	review, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Review])
	if err != nil {
		return nil, fmt.Errorf("GetExistingReviewCombination: %w", err)
	}

	return &review, nil
}

func (r *reviewRepository) GetReviewForBooking(ctx context.Context, userId uuid.UUID, bookingId int) (*domain.Review, error) {
	query := `
	select r.id, r.booking_id, r.user_id, r.merchant_id, r.location_id, r.employee_id, r.service_id, r.service_name, r.rating, 
	r.comment, r.status, r.report_reason, r.moderated_at, r.created_at, r.updated_at
	from "Booking" b
	join "Review" r on (
		r.user_id = $1
		and (
			r.booking_id = b.id
			or (
				r.status <> 'removed' and b.employee_id is not null and b.service_id is not null and r.merchant_id = b.merchant_id
				and r.location_id = b.location_id and r.employee_id = b.employee_id and r.service_id = b.service_id
			)
		)
	)
	where b.id = $2
	order by (r.booking_id = b.id) desc, (r.status <> 'removed') desc, r.created_at desc
	limit 1
	`

	rows, _ := r.db.Query(ctx, query, userId, bookingId)
	review, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Review])
	if err != nil {
		return nil, fmt.Errorf("GetReviewForBooking: %w", err)
	}

	return &review, nil
}
