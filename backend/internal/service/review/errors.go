package review

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrBookingNotFound:           http.StatusNotFound,
	ErrBookingNotCompleted:       http.StatusBadRequest,
	ErrBookingExpiredForReview:   http.StatusBadRequest,
	ErrSubsequentBookingRequired: http.StatusBadRequest,

	ErrBookingAlreadyReviewed: http.StatusConflict,
	ErrReviewAlreadyExists:    http.StatusConflict,

	ErrReviewFrozen:  http.StatusForbidden,
	ErrReviewRemoved: http.StatusForbidden,
	ErrUnauthorized:  http.StatusForbidden,
}

var (
	ErrBookingNotFound           = &apperr.Error{Code: "booking_not_found", Message: "booking not found"}
	ErrBookingNotCompleted       = &apperr.Error{Code: "booking_not_completed", Message: "only completed bookings can be reviewed"}
	ErrBookingExpiredForReview   = &apperr.Error{Code: "booking_expired_for_review", Message: "bookings older than 15 days cannot be reviewed"}
	ErrBookingAlreadyReviewed    = &apperr.Error{Code: "booking_already_reviewed", Message: "this booking has already been reviewed"}
	ErrSubsequentBookingRequired = &apperr.Error{Code: "subsequent_booking_required", Message: "a new booking after the previous review date is required to leave another review"}
	ErrReviewAlreadyExists       = &apperr.Error{Code: "review_already_exists", Message: "this service and team member combination has already been reviewed"}
	ErrReviewFrozen              = &apperr.Error{Code: "review_frozen", Message: "review is currently under moderation review and cannot be edited"}
	ErrReviewRemoved             = &apperr.Error{Code: "review_removed", Message: "removed reviews cannot be edited"}
	ErrUnauthorized              = &apperr.Error{Code: "unauthorized", Message: "you are not authorized to perform this action"}
)
