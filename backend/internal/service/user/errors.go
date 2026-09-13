package user

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	domain.ErrEmailNotUnique:       http.StatusConflict,
	domain.ErrPhoneNumberNotUnique: http.StatusConflict,
}
