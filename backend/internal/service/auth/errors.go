package auth

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrIncorrectEmailOrPassword:    http.StatusUnauthorized,
	domain.ErrEmailNotUnique:       http.StatusConflict,
	domain.ErrPhoneNumberNotUnique: http.StatusConflict,
}

var ErrIncorrectEmailOrPassword = &apperr.Error{Code: "incorrect_email_or_password", Message: "incorrect email or password"}
