package team

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrPreferencesForbidden: http.StatusForbidden,
}

var ErrPreferencesForbidden = &apperr.Error{Code: "preferences_forbidden", Message: "you do not have permission to access these preferences"}
