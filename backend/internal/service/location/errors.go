package location

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrLocationNotFound: http.StatusNotFound,
}

var ErrLocationNotFound = &apperr.Error{Code: "location_not_found", Message: "location not found"}
