package media

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrImageNotFound:             http.StatusNotFound,
	ErrInvalidUploadSize:         http.StatusBadRequest,
	ErrUploadMissing:             http.StatusConflict,
	ErrUploadAlreadyCompleted:    http.StatusConflict,
	ErrUploadInvalid:             http.StatusBadRequest,
	ErrEmployeeNotFound:          http.StatusNotFound,
	ErrEmployeeAvatarForbidden:   http.StatusForbidden,
	ErrLocationNotFound:          http.StatusNotFound,
	ErrLocationImageNotFound:     http.StatusNotFound,
	ErrLocationImageLimitReached: http.StatusConflict,
	ErrInvalidImageOrder:         http.StatusBadRequest,
}

var ErrImageNotFound = &apperr.Error{Code: "image_not_found", Message: "image not found for this id"}
var ErrInvalidUploadSize = &apperr.Error{Code: "invalid_upload_size", Message: "the image must be smaller than 10 MB"}
var ErrUploadMissing = &apperr.Error{Code: "upload_missing", Message: "the image has not been uploaded yet"}
var ErrUploadAlreadyCompleted = &apperr.Error{Code: "upload_already_completed", Message: "the upload has already been completed"}
var ErrUploadInvalid = &apperr.Error{Code: "upload_invalid", Message: "the uploaded file does not match the requested upload"}
var ErrEmployeeNotFound = &apperr.Error{Code: "employee_not_found", Message: "employee not found for this id"}
var ErrEmployeeAvatarForbidden = &apperr.Error{Code: "employee_avatar_forbidden", Message: "you can only change your own profile picture"}
var ErrLocationNotFound = &apperr.Error{Code: "location_not_found", Message: "location not found for this id"}
var ErrLocationImageNotFound = &apperr.Error{Code: "location_image_not_found", Message: "location image not found for this id"}
var ErrLocationImageLimitReached = &apperr.Error{Code: "location_image_limit_reached", Message: "the location has reached the maximum number of images for your subscription"}
var ErrInvalidImageOrder = &apperr.Error{Code: "invalid_image_order", Message: "the order must contain every image of the location exactly once"}
