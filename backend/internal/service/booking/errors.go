package booking

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrCustomerIsBlacklisted: http.StatusForbidden,
	ErrTimeIsNotAvailable:    http.StatusConflict,
}

var ErrCustomerIsBlacklisted = &apperr.Error{Code: "customer_is_blacklisted", Message: "you are blacklisted, please contact the merchant by email or phone to make a booking"}
var ErrTimeIsNotAvailable = &apperr.Error{Code: "time_is_not_available", Message: "the selected time is not available, please try again"}
