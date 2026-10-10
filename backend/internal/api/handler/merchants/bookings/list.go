package bookings

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type bookingListItemResp struct {
	Id                int                      `json:"id"`
	BookingType       types.BookingType        `json:"booking_type"`
	IsRecurring       bool                     `json:"is_recurring"`
	FromDate          time.Time                `json:"from_date"`
	ToDate            time.Time                `json:"to_date"`
	ServiceName       string                   `json:"service_name"`
	ServiceColor      *string                  `json:"service_color"`
	FormattedLocation string                   `json:"formatted_location"`
	Price             currencyx.FormattedPrice `json:"price"`
	PriceType         types.PriceType          `json:"price_type"`
	BookingStatus     types.BookingStatus      `json:"booking_status"`
	ParticipantStatus *types.BookingStatus     `json:"participant_status"`
	EmployeeFirstName *string                  `json:"employee_first_name"`
	EmployeeLastName  *string                  `json:"employee_last_name"`
}

type getBookingsResp struct {
	Bookings    []bookingListItemResp `json:"bookings"`
	HasNextPage bool                  `json:"has_next_page"`
	NextCursor  *string               `json:"next_cursor"`
}

func (h *Handler) GetBookings(w http.ResponseWriter, r *http.Request) error {
	var customerId *uuid.UUID
	if value := r.URL.Query().Get("customer_id"); value != "" {
		parsed, err := uuid.Parse(value)
		if err != nil {
			return validate.NewError("invalid customer_id query parameter")
		}

		customerId = &parsed
	}

	var employeeId *int
	if value := r.URL.Query().Get("employee_id"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			return validate.NewError("invalid employee_id query parameter")
		}

		employeeId = &parsed
	}

	if (customerId == nil) == (employeeId == nil) {
		return validate.NewError("exactly one of customer_id or employee_id is required")
	}

	var statuses []types.BookingStatus

	statusValues := strings.Split(r.URL.Query().Get("status"), ",")
	if len(statusValues) == 1 && statusValues[0] == "all" {
		statuses = []types.BookingStatus{
			types.BookingStatusBooked,
			types.BookingStatusConfirmed,
			types.BookingStatusCompleted,
			types.BookingStatusCancelled,
			types.BookingStatusNoShow,
		}
	} else {
		statuses = make([]types.BookingStatus, 0, len(statusValues))
		seenStatuses := make(map[types.BookingStatus]bool, len(statusValues))

		for _, value := range statusValues {
			status, err := types.NewBookingStatus(value)
			if err != nil || seenStatuses[status] {
				return validate.NewError("invalid status query parameter")
			}

			seenStatuses[status] = true
			statuses = append(statuses, status)
		}
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		return validate.NewError("invalid limit query parameter")
	}

	if limit < 1 || limit > 20 {
		return validate.NewError("limit must be between 1 and 20")
	}

	var before *time.Time
	if value := r.URL.Query().Get("before"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return validate.NewError("invalid before query parameter")
		}

		before = &parsed
	}

	result, err := h.service.GetBookings(r.Context(), bookingServ.GetBookingsInput{
		CustomerId: customerId,
		EmployeeId: employeeId,
		Statuses:   statuses,
		Before:     before,
		Cursor:     r.URL.Query().Get("cursor"),
		PageSize:   limit,
	})
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "GetBookings")
	}

	httputil.Success(w, http.StatusOK, mapToGetBookingsResp(result))

	return nil
}
