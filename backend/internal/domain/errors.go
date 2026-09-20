package domain

import "github.com/miketsu-inc/reservations/backend/pkg/apperr"

var ErrEmailNotUnique = &apperr.Error{Code: "email_not_unique", Message: "this email is already used"}
var ErrPhoneNumberNotUnique = &apperr.Error{Code: "phone_number_not_unique", Message: "this phone number is already used"}

var ErrBookingNotFound = &apperr.Error{Code: "booking_not_found", Message: "booking not found"}
