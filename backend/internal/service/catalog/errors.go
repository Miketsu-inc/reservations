package catalog

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
)

var ErrStatus = apperr.StatusMap{
	ErrServicePhasesRequired:                  http.StatusBadRequest,
	ErrGroupServiceRequiresSinglePhase:        http.StatusBadRequest,
	ErrGroupServiceRequiresMaxParticipants:    http.StatusBadRequest,
	ErrServicePriceCurrencyMismatch:           http.StatusBadRequest,
	domain.ErrEmployeesRequired:               http.StatusBadRequest,
	domain.ErrDuplicateEmployee:               http.StatusBadRequest,
	domain.ErrEmployeeNotActive:               http.StatusBadRequest,
	ErrEmployeeServiceDurationOutOfRange:      http.StatusBadRequest,
	ErrEmployeeServicePriceCurrencyMismatch:   http.StatusBadRequest,
	ErrEmployeeServicePriceOutOfRange:         http.StatusBadRequest,
	ErrEmployeeServiceParticipantsOutOfRange:  http.StatusBadRequest,
	ErrEmployeeServiceParticipantRangeInvalid: http.StatusBadRequest,
	ErrEmployeeServiceBufferTimeOutOfRange:    http.StatusBadRequest,
	ErrInvalidEmployeeServicePhase:            http.StatusBadRequest,
	ErrDuplicateEmployeeServicePhase:          http.StatusBadRequest,
	ErrDuplicateServiceId:                     http.StatusBadRequest,
	ErrDuplicateServiceCategoryId:             http.StatusBadRequest,
}

var (
	ErrDuplicateServiceId           = &apperr.Error{Code: "duplicate_service_id", Message: "a service can only appear once"}
	ErrServicePhasesRequired        = &apperr.Error{Code: "service_phases_required", Message: "service must have at least one phase"}
	ErrServicePriceCurrencyMismatch = &apperr.Error{Code: "service_price_currency_mismatch", Message: "service price currency must match the merchant currency"}

	ErrGroupServiceRequiresSinglePhase     = &apperr.Error{Code: "group_service_requires_single_phase", Message: "group services must have exactly one phase"}
	ErrGroupServiceRequiresMaxParticipants = &apperr.Error{Code: "group_service_requires_max_participants", Message: "group services must have a maximum participant count"}

	ErrEmployeeServiceDurationOutOfRange      = &apperr.Error{Code: "employee_service_duration_out_of_range", Message: "team member service duration must be between 1 and 1440 minutes"}
	ErrEmployeeServicePriceCurrencyMismatch   = &apperr.Error{Code: "employee_service_price_currency_mismatch", Message: "team member service price currency must match the merchant currency"}
	ErrEmployeeServicePriceOutOfRange         = &apperr.Error{Code: "employee_service_price_out_of_range", Message: "team member service price must be between 0 and 1000000"}
	ErrEmployeeServiceParticipantsOutOfRange  = &apperr.Error{Code: "employee_service_participants_out_of_range", Message: "team member participant limits must be at least 1"}
	ErrEmployeeServiceParticipantRangeInvalid = &apperr.Error{Code: "employee_service_participant_range_invalid", Message: "team member minimum participants cannot exceed maximum participants"}
	ErrEmployeeServiceBufferTimeOutOfRange    = &apperr.Error{Code: "employee_service_buffer_time_out_of_range", Message: "team member buffer time must be between 0 and 1440 minutes"}
	ErrInvalidEmployeeServicePhase            = &apperr.Error{Code: "invalid_employee_service_phase", Message: "team member phase does not belong to this service"}
	ErrDuplicateEmployeeServicePhase          = &apperr.Error{Code: "duplicate_employee_service_phase", Message: "a service phase can only be overridden once per team member"}
	ErrDuplicateServiceCategoryId             = &apperr.Error{Code: "duplicate_service_category_id", Message: "a service category can only appear once"}
)
