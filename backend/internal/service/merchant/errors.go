package merchant

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrMerchantUrlNotUnique: http.StatusConflict,
}

var ErrMerchantUrlNotUnique = &apperr.Error{Code: "merchant_url_not_unique", Message: "this merchant url is already used"}
