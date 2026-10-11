package location

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrLocationNotFound:      http.StatusNotFound,
	ErrLocationAlreadyExists: http.StatusConflict,
}

var ErrLocationNotFound = &apperr.Error{Code: "location_not_found", Message: "location not found"}
var ErrLocationAlreadyExists = &apperr.Error{Code: "location_already_exists", Message: "this merchant already has a location"}
