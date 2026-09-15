package bookings

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type getDashboardBookingsResp struct {
	Bookings []dashboardBookingResp `json:"bookings"`
}

type dashboardBookingResp struct {
	ID                  int                      `json:"id"`
	BookingType         types.BookingType        `json:"booking_type"`
	BookingStatus       types.BookingStatus      `json:"booking_status"`
	ParticipantStatus   *types.BookingStatus     `json:"participant_status"`
	IsRecurring         bool                     `json:"is_recurring"`
	FromDate            time.Time                `json:"from_date"`
	ToDate              time.Time                `json:"to_date"`
	CustomerNote        *string                  `json:"customer_note"`
	MerchantNote        *string                  `json:"merchant_note"`
	ServiceName         string                   `json:"service_name"`
	ServiceColor        *string                  `json:"service_color"`
	Price               currencyx.FormattedPrice `json:"price"`
	PriceType           types.PriceType          `json:"price_type"`
	CurrentParticipants int                      `json:"current_participants"`
	MaxParticipants     int                      `json:"max_participants"`
	CustomerFirstName   *string                  `json:"customer_first_name"`
	CustomerLastName    *string                  `json:"customer_last_name"`
	EmployeeFirstName   *string                  `json:"employee_first_name"`
	EmployeeLastName    *string                  `json:"employee_last_name"`
}

func (h *Handler) GetDashboardBookings(w http.ResponseWriter, r *http.Request) error {
	view := chi.URLParam(r, "view")
	if view != "latest" && view != "upcoming" {
		return validate.NewError("view must be either latest or upcoming")
	}

	result, err := h.service.GetDashboardBookings(r.Context(), view)
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "GetDashboardBookings")
	}

	httputil.Success(w, http.StatusOK, mapToGetDashboardBookingsResp(result))

	return nil
}
