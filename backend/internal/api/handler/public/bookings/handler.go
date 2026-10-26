package bookings

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service    *bookingServ.Service
	middleware *middleware.Manager
}

func NewHandler(s *bookingServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, middleware: m}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.JwtAuthentication)

		r.Post("/", h.CreateByCustomer)
	})

	return r
}

type createBookingByCustomerReq struct {
	MerchantName string `json:"merchant_name" validate:"required"`
	ServiceId    int    `json:"service_id" validate:"required"`
	LocationId   int    `json:"location_id" validate:"required"`
	StartsAt     string `json:"starts_at" validate:"required"`
	CustomerNote string `json:"customer_note"`
	// nil when no-preference is selected for 1-on-1 bookings
	EmployeeId *int `json:"employee_id"`
	// only present on group bookings
	BookingId *int `json:"booking_id"`
}

func (h *Handler) CreateByCustomer(w http.ResponseWriter, r *http.Request) error {
	var req createBookingByCustomerReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	input, err := mapToCreateByCustomerInput(req)
	if err != nil {
		return validate.NewError(err.Error())
	}

	err = h.service.CreateByCustomer(r.Context(), input)
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "CreateByCustomer")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}
