package bookings

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
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

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(h.middleware.JwtAuthentication)
		r.Use(h.middleware.Language)

		r.Post("/", h.CreateByCustomer)
		r.Delete("/{id}", h.CancelByCustomer)
		r.Get("/{id}", h.GetByCustomer)
	})

	return r
}

type createBookingByCustomerReq struct {
	MerchantName string `json:"merchant_name" validate:"required"`
	ServiceId    int    `json:"service_id" validate:"required"`
	LocationId   int    `json:"location_id" validate:"required"`
	TimeStamp    string `json:"timeStamp" validate:"required"`
	CustomerNote string `json:"customer_note"`
	// only present on group bookings
	BookingId *int `json:"booking_id"`
}

func (h *Handler) CreateByCustomer(w http.ResponseWriter, r *http.Request) {
	var req createBookingByCustomerReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.CreateByCustomer(r.Context(), mapToCreateByCustomerInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

type cancelByCustomerReq struct {
	BookingId    int    `json:"booking_id" validate:"required"`
	MerchantName string `json:"merchant_name" validate:"required"`
}

func (h *Handler) CancelByCustomer(w http.ResponseWriter, r *http.Request) {
	var req cancelByCustomerReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid booking id: %w", err))
		return
	}

	err = h.service.CancelByCustomer(r.Context(), urlId, mapToCancelByCustomerInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getByCustomerResp struct {
	FromDate          time.Time                `json:"from_date"`
	ToDate            time.Time                `json:"to_date"`
	ServiceName       string                   `json:"service_name"`
	CancelDeadline    int                      `json:"cancel_deadline"`
	FormattedLocation string                   `json:"formatted_location"`
	Price             currencyx.FormattedPrice `json:"price"`
	PriceType         types.PriceType          `json:"price_type"`
	MerchantName      string                   `json:"merchant_name"`
	Status            types.BookingStatus      `json:"status"`
}

func (h *Handler) GetByCustomer(w http.ResponseWriter, r *http.Request) {
	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid booking id: %w", err))
		return
	}

	publicBooking, err := h.service.GetByCustomer(r.Context(), urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetByCustomerResp(publicBooking))
}
