package db

import (
	"context"
	"fmt"
	"strings"
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

func (r *reviewRepository) DeleteReply(ctx context.Context, merchantId uuid.UUID, reviewId int) error {
	query := `
	delete from "ReviewReply" rr
	using "Review" r
	where rr.review_id = r.id and r.id = $1 and r.merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, reviewId, merchantId)
	if err != nil {
		return fmt.Errorf("DeleteReply: %w", err)
	}

	return nil
}

func (r *reviewRepository) UpsertReply(ctx context.Context, merchantId uuid.UUID, reviewId int, employeeId int, comment string) error {
	query := `
	insert into "ReviewReply" (review_id, employee_id, comment, created_at, updated_at) 
	select r.id, $3, $4, now(), now()
	from "Review" r
	where r.id = $1 and r.merchant_id = $2
	on conflict (review_id) do update 
	set employee_id = excluded.employee_id, comment = excluded.comment, updated_at = now()
	`

	_, err := r.db.Exec(ctx, query, reviewId, merchantId, employeeId, comment)
	if err != nil {
		return fmt.Errorf("UpsertReply: %w", err)
	}

	return nil
}

func (r *reviewRepository) GetReviewStats(ctx context.Context, merchantId uuid.UUID, locationId *int) (domain.ReviewStats, error) {
	query := `
	select 
		coalesce(round(avg(r.rating)::numeric, 2), 0)::float8 as average_rating,
		count(*)::int as total_reviews,
		count(*) filter (where r.rating = 5)::int as count_5_star,
		count(*) filter (where r.rating = 4)::int as count_4_star,
		count(*) filter (where r.rating = 3)::int as count_3_star,
		count(*) filter (where r.rating = 2)::int as count_2_star,
		count(*) filter (where r.rating = 1)::int as count_1_star
	from "Review" r
	where r.merchant_id = $1 and r.status <> 'removed'
	  and ($2::int is null or r.location_id = $2)
	`

	rows, err := r.db.Query(ctx, query, merchantId, locationId)
	if err != nil {
		return domain.ReviewStats{}, fmt.Errorf("GetMerchantReviewStats query: %w", err)
	}
	defer rows.Close()

	stats, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.ReviewStats])
	if err != nil {
		return domain.ReviewStats{}, fmt.Errorf("GetMerchantReviewStats collect: %w", err)
	}

	return stats, nil
}

var reviewOrderBy = map[string]string{
	"newest":        "r.created_at desc, r.id desc",
	"oldest":        "r.created_at asc, r.id asc",
	"highest":       "r.rating desc, r.created_at DESC, r.id desc",
	"lowest":        "r.rating asc, r.created_at DESC, r.id desc",
	"most_relevant": "length(coalesce(r.comment, '')) desc, r.created_at desc, r.id desc",
}

func (r *reviewRepository) GetReviews(ctx context.Context, filter domain.ReviewFilter) ([]domain.ReviewDetails, error) {
	var conditions []string
	var args []any
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("r.merchant_id = $%d", argIdx))
	args = append(args, filter.MerchantId)
	argIdx++

	if filter.LocationId != nil {
		conditions = append(conditions, fmt.Sprintf("r.location_id = $%d", argIdx))
		args = append(args, *filter.LocationId)
		argIdx++
	}

	if len(filter.Statuses) > 0 {
		conditions = append(conditions, fmt.Sprintf("r.status = any($%d)", argIdx))
		args = append(args, filter.Statuses)
		argIdx++
	} else {
		conditions = append(conditions, "r.status <> 'removed'")
	}

	if filter.Rating != nil {
		conditions = append(conditions, fmt.Sprintf("r.rating = $%d", argIdx))
		args = append(args, *filter.Rating)
		argIdx++
	}

	if len(filter.EmployeeIds) > 0 {
		conditions = append(conditions, fmt.Sprintf("r.employee_id = any($%d)", argIdx))
		args = append(args, filter.EmployeeIds)
		argIdx++
	}

	if len(filter.ServiceIds) > 0 {
		conditions = append(conditions, fmt.Sprintf("r.service_id = any($%d)", argIdx))
		args = append(args, filter.ServiceIds)
		argIdx++
	}

	if filter.HasReply != nil {
		if *filter.HasReply {
			conditions = append(conditions, "rr.review_id is not null")
		} else {
			conditions = append(conditions, "rr.review_id is null")
		}
	}

	orderBy, ok := reviewOrderBy[filter.Sort]
	if !ok {
		orderBy = reviewOrderBy["newest"]
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	limitArg := argIdx
	args = append(args, limit)
	argIdx++

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	offsetArg := argIdx
	args = append(args, offset)
	argIdx++

	query := fmt.Sprintf(`
	SELECT 
		r.id, r.booking_id, r.user_id, r.merchant_id, r.location_id, r.employee_id, r.service_id, r.service_name, r.rating, r.comment, r.status, 
		r.report_reason, r.moderated_at, r.created_at, r.updated_at,
		jsonb_build_object(
			'id', u.id,
			'first_name', u.first_name,
			'last_name', u.last_name,
			'avatar_image_id', u.avatar_image_id
		) as customer,
		case when r.employee_id is not null then
			jsonb_build_object(
				'id', r.employee_id,
				'first_name', coalesce(e.first_name, eu.first_name),
				'last_name', coalesce(e.last_name, eu.last_name),
				'avatar_image_id', coalesce(e.avatar_image_id, eu.avatar_image_id)
			)
		else null end as employee,
		case when rr.review_id is not null then
			jsonb_build_object(
				'comment', rr.comment,
				'created_at', rr.created_at,
				'updated_at', rr.updated_at,
				'employee', case when rr.employee_id is not null then
					jsonb_build_object(
						'id', rr.employee_id,
						'first_name', coalesce(re.first_name, reu.first_name),
						'last_name', coalesce(re.last_name, reu.last_name),
						'avatar_image_id', coalesce(re.avatar_image_id, reu.avatar_image_id)
					)
				else null end
			)
		else null end as reply
	from "Review" r
	join "User" u ON u.id = r.user_id
	left join "Employee" e ON e.id = r.employee_id
	left join "User" eu ON eu.id = e.user_id
	left join "ReviewReply" rr ON rr.review_id = r.id
	left join "Employee" re ON re.id = rr.employee_id
	left join "User" reu ON reu.id = re.user_id
	where %s
	order by %s
	limit $%d offset $%d
	`, strings.Join(conditions, " and "), orderBy, limitArg, offsetArg)

	rows, _ := r.db.Query(ctx, query, args...)
	reviews, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ReviewDetails])
	if err != nil {
		return nil, fmt.Errorf("GetReviews: %w", err)
	}

	return reviews, nil
}
