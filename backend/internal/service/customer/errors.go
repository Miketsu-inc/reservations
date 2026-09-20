package customer

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrCustomerTransferConflict: http.StatusConflict,
}

var ErrCustomerTransferConflict = &apperr.Error{
	Code:    "customer_transfer_conflict",
	Message: "cannot transfer bookings because both customers participate in the same booking or recurring series",
}
