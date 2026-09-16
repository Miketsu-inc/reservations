package auth

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrIncorrectEmailOrPassword:    http.StatusUnauthorized,
	ErrOauthAccountAlreadyExists:   http.StatusConflict,
	ErrOauthEmailUnavailable:       http.StatusBadRequest,
	domain.ErrEmailNotUnique:       http.StatusConflict,
	domain.ErrPhoneNumberNotUnique: http.StatusConflict,
}

var ErrIncorrectEmailOrPassword = &apperr.Error{Code: "incorrect_email_or_password", Message: "incorrect email or password"}
var ErrOauthAccountAlreadyExists = &apperr.Error{Code: "oauth_account_already_exists", Message: "account with this email already exists"}
var ErrOauthEmailUnavailable = &apperr.Error{Code: "oauth_email_unavailable", Message: "oauth provider did not return a usable email address"}
