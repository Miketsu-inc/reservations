package catalog

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrServicePhasesRequired:                http.StatusBadRequest,
	ErrGroupServiceRequiresSinglePhase:      http.StatusBadRequest,
	ErrGroupServiceRequiresMaxParticipants:  http.StatusBadRequest,
	ErrServicePriceCurrencyMismatch:         http.StatusBadRequest,
	ErrServiceRequiresEmployee:              http.StatusBadRequest,
	ErrDuplicateServiceEmployee:             http.StatusBadRequest,
	ErrEmployeeServiceDurationOutOfRange:    http.StatusBadRequest,
	ErrEmployeeServicePriceCurrencyMismatch: http.StatusBadRequest,
	ErrEmployeeServicePriceOutOfRange:       http.StatusBadRequest,
	ErrEmployeeNotActiveForMerchant:         http.StatusBadRequest,
	ErrDuplicateServiceId:                   http.StatusBadRequest,
	ErrDuplicateServiceCategoryId:           http.StatusBadRequest,
}

var (
	ErrDuplicateServiceId           = &apperr.Error{Code: "duplicate_service_id", Message: "a service can only appear once"}
	ErrServicePhasesRequired        = &apperr.Error{Code: "service_phases_required", Message: "service must have at least one phase"}
	ErrServiceRequiresEmployee      = &apperr.Error{Code: "service_requires_employee", Message: "at least one team member must be assigned to the service"}
	ErrServicePriceCurrencyMismatch = &apperr.Error{Code: "service_price_currency_mismatch", Message: "service price currency must match the merchant currency"}

	ErrGroupServiceRequiresSinglePhase     = &apperr.Error{Code: "group_service_requires_single_phase", Message: "group services must have exactly one phase"}
	ErrGroupServiceRequiresMaxParticipants = &apperr.Error{Code: "group_service_requires_max_participants", Message: "group services must have a maximum participant count"}

	ErrDuplicateServiceEmployee             = &apperr.Error{Code: "duplicate_service_employee", Message: "a team member can only be assigned to a service once"}
	ErrEmployeeServiceDurationOutOfRange    = &apperr.Error{Code: "employee_service_duration_out_of_range", Message: "team member service duration must be between 1 and 1440 minutes"}
	ErrEmployeeServicePriceCurrencyMismatch = &apperr.Error{Code: "employee_service_price_currency_mismatch", Message: "team member service price currency must match the merchant currency"}
	ErrEmployeeServicePriceOutOfRange       = &apperr.Error{Code: "employee_service_price_out_of_range", Message: "team member service price must be between 0 and 1000000"}
	ErrEmployeeNotActiveForMerchant         = &apperr.Error{Code: "employee_not_active_for_merchant", Message: "team member must be active and belong to this merchant"}

	ErrDuplicateServiceCategoryId = &apperr.Error{Code: "duplicate_service_category_id", Message: "a service category can only appear once"}
)
