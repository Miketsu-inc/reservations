package booking

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrCustomerIsBlacklisted: http.StatusForbidden,
	ErrTimeIsNotAvailable:    http.StatusConflict,

	ErrEmployeeNotFoundForService: http.StatusNotFound,
	ErrBookingInactiveService:     http.StatusConflict,
}

var ErrCustomerIsBlacklisted = &apperr.Error{Code: "customer_is_blacklisted", Message: "you are blacklisted, please contact the merchant by email or phone to make a booking"}
var ErrTimeIsNotAvailable = &apperr.Error{Code: "time_is_not_available", Message: "the selected time is not available, please try again"}
var ErrEmployeeNotFoundForService = &apperr.Error{Code: "employee_not_found_for_service", Message: "there is no employee doing this service"}
var ErrBookingInactiveService = &apperr.Error{Code: "booking_inactive_service", Message: "cannot book to an inactive service"}
