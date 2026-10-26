package booking

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrCustomerIsBlacklisted: http.StatusForbidden,
	ErrTimeIsNotAvailable:    http.StatusConflict,

	ErrEmployeeNotFoundForService: http.StatusNotFound,
	ErrBookingNotFound:            http.StatusNotFound,
	ErrBookingInactiveService:      http.StatusConflict,
	ErrBookingCancelDeadlinePassed: http.StatusConflict,
	ErrBookingCannotBeModified:     http.StatusConflict,
}

var ErrCustomerIsBlacklisted = &apperr.Error{Code: "customer_is_blacklisted", Message: "you are blacklisted, please contact the merchant by email or phone to make a booking"}
var ErrTimeIsNotAvailable = &apperr.Error{Code: "time_is_not_available", Message: "the selected time is not available, please try again"}
var ErrBookingNotFound = &apperr.Error{Code: "booking_not_found", Message: "booking not found"}
var ErrBookingCancelDeadlinePassed = &apperr.Error{Code: "booking_cancel_deadline_passed", Message: "it's too late to cancel this booking"}
var ErrBookingCannotBeModified = &apperr.Error{Code: "booking_cannot_be_modified", Message: "booking has already been cancelled"}
var ErrEmployeeNotFoundForService = &apperr.Error{Code: "employee_not_found_for_service", Message: "there is no employee doing this service"}
var ErrBookingInactiveService = &apperr.Error{Code: "booking_inactive_service", Message: "cannot book to an inactive service"}
