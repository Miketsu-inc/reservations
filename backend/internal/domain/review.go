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

	UpsertReply(ctx context.Context, merchantId uuid.UUID, reviewId int, employeeId int, comment string) error
	DeleteReply(ctx context.Context, merchantId uuid.UUID, reviewId int) error

	GetReviews(ctx context.Context, filter ReviewFilter) ([]ReviewDetails, error)
	GetReviewStats(ctx context.Context, merchantId uuid.UUID, locationId *int) (ReviewStats, error)
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

type ReviewReply struct {
	ReviewId   int
	EmployeeId *int
	Comment    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ReviewFilter struct {
	MerchantId  uuid.UUID
	LocationId  *int
	Rating      *int16
	EmployeeIds []int
	ServiceIds  []int
	HasReply    *bool
	Statuses    []string
	Sort        string
	Limit       int
	Offset      int
}

type ReviewStats struct {
	AverageRating float64 `db:"average_rating"`
	TotalReviews  int     `db:"total_reviews"`
	Count5Star    int     `db:"count_5_star"`
	Count4Star    int     `db:"count_4_star"`
	Count3Star    int     `db:"count_3_star"`
	Count2Star    int     `db:"count_2_star"`
	Count1Star    int     `db:"count_1_star"`
}

type ReviewReplyInfo struct {
	Comment   string               `json:"comment"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	Employee  *MinimalEmployeeInfo `json:"employee,omitempty"`
}

type ReviewDetails struct {
	Review
	Customer MinimalCustomerInfo  `db:"customer"`
	Employee *MinimalEmployeeInfo `db:"employee"`
	Reply    *ReviewReplyInfo     `db:"reply"`
}
