package blockedtime

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrBlockedTimeNotFound:           http.StatusNotFound,
	ErrAllDayBlockedTimeDateRequired: http.StatusBadRequest,
	ErrTimedBlockedTimeDatesRequired: http.StatusBadRequest,
	ErrInvalidBlockedTimeDateRange:   http.StatusBadRequest,
	ErrBlockedTimeDurationTooLong:    http.StatusBadRequest,
}

var ErrBlockedTimeNotFound = &apperr.Error{Code: "blocked_time_not_found", Message: "blocked time not found"}
var ErrAllDayBlockedTimeDateRequired = &apperr.Error{Code: "all_day_blocked_time_date_required", Message: "an all-day blocked time requires blocked_day only"}
var ErrTimedBlockedTimeDatesRequired = &apperr.Error{Code: "timed_blocked_time_dates_required", Message: "a timed blocked time requires from_date and to_date only"}
var ErrInvalidBlockedTimeDateRange = &apperr.Error{Code: "invalid_blocked_time_date_range", Message: "to_date must be after from_date"}
var ErrBlockedTimeDurationTooLong = &apperr.Error{Code: "blocked_time_duration_too_long", Message: "blocked time duration must not exceed 24 hours"}
