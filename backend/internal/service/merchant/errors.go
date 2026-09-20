package merchant

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrMerchantUrlNotUnique: http.StatusConflict,
	ErrMerchantNotFound:     http.StatusNotFound,
	ErrBookingNotFound:      http.StatusNotFound,

	ErrMerchantServiceMismatch: http.StatusForbidden,
	ErrInvalidBookingType:      http.StatusForbidden,
}

var ErrMerchantUrlNotUnique = &apperr.Error{Code: "merchant_url_not_unique", Message: "this merchant url is already used"}
var ErrMerchantNotFound = &apperr.Error{Code: "merchant_not_found", Message: "merchant not found"}
var ErrBookingNotFound = &apperr.Error{Code: "booking_not_found", Message: "booking not found"}

var ErrMerchantServiceMismatch = &apperr.Error{Code: "merchant_service_mismatch", Message: "this service does not belong to this merchant"}
var ErrInvalidBookingType = &apperr.Error{Code: "invalid_booking_type", Message: "this service has invalid type"}
