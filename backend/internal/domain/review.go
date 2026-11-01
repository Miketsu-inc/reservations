package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type ReviewRepository interface {
	WithTx(tx db.DBTX) ReviewRepository

	NewReview(ctx context.Context, review Review) error
	UpdateReview(ctx context.Context, reviewId int, userId uuid.UUID, bookingId *int, rating int16, comment *string) error
	DeleteReview(ctx context.Context, reviewId int, userId uuid.UUID) error
	ReportReview(ctx context.Context, reviewId int, merchantId uuid.UUID, reportReason string) error
	RemoveReview(ctx context.Context, reviewId int, moderatedAt time.Time) error

	GetReviewById(ctx context.Context, reviewId int) (Review, error)
	GetExistingReviewCombination(ctx context.Context, userId uuid.UUID, merchantId uuid.UUID, locationId int, employeeId int, serviceId int) (*Review, error)
	GetReviewForBooking(ctx context.Context, userId uuid.UUID, bookingId int) (*Review, error)
}

type Review struct {
	Id           int        `db:"id"`
	BookingId    *int       `db:"booking_id"`
	UserId       uuid.UUID  `db:"user_id"`
	MerchantId   uuid.UUID  `db:"merchant_id"`
	LocationId   int        `db:"location_id"`
	EmployeeId   *int       `db:"employee_id"`
	ServiceId    *int       `db:"service_id"`
	ServiceName  string     `db:"service_name"`
	Rating       int16      `db:"rating"`
	Comment      *string    `db:"comment"`
	Status       string     `db:"status"`
	ReportReason *string    `db:"report_reason"`
	ModeratedAt  *time.Time `db:"moderated_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}
