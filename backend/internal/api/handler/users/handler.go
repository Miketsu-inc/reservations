package users

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	authServ "github.com/miketsu-inc/reservations/backend/internal/service/auth"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	userServ "github.com/miketsu-inc/reservations/backend/internal/service/user"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service     *userServ.Service
	bookingServ *bookingServ.Service
	authServ    *authServ.Service
	middleware  *middleware.Manager
}

func NewHandler(s *userServ.Service, b *bookingServ.Service, a *authServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, bookingServ: b, authServ: a, middleware: m}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(h.middleware.JwtAuthentication)
		r.Use(h.middleware.Language)

		r.Put("/", h.Edit)
		r.Delete("/", h.Delete)

		r.Get("/bookings", h.GetBookings)
		r.Put("/password", h.UpdatePassword)
	})

	return r
}

type editReq struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	Email       string `json:"email" validate:"required,email"`
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	var req editReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.Edit(r.Context(), mapToEditInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.service.Delete(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	jwt.DeleteJwts(w)
}

type getBookingsResp struct {
	Bookings    []bookingForUser `json:"bookings"`
	HasNextpage bool             `json:"has_next_page"`
	NextCursor  *string          `json:"next_cursor"`
}

type bookingForUser struct {
	Id                int                      `json:"id"`
	Status            string                   `json:"status"`
	BookingType       string                   `json:"booking_type"`
	IsRecurring       bool                     `json:"is_recurring"`
	FromDate          time.Time                `json:"from_date"`
	ToDate            time.Time                `json:"to_date"`
	Price             currencyx.FormattedPrice `json:"price"`
	MerchantName      string                   `json:"merchant_name"`
	MerchantUrl       string                   `json:"merchant_url"`
	FormattedLocation string                   `json:"formatted_location"`
	ServiceName       string                   `json:"service_name"`
	EmployeeFirstName *string                  `json:"employee_first_name"`
	EmployeeLastName  *string                  `json:"employee_last_name"`
}

func (h *Handler) GetBookings(w http.ResponseWriter, r *http.Request) {
	urlStatus := r.URL.Query().Get("status")
	if urlStatus != "upcoming" && urlStatus != "completed" && urlStatus != "cancelled" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid status query parameter"))
		return
	}

	urlCursor := r.URL.Query().Get("cursor")

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid limit query parameter"))
		return
	}

	if limit > 10 {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("limit cannot be higher than 10"))
		return
	}

	bookings, err := h.bookingServ.GetForUser(r.Context(), urlStatus, urlCursor, limit)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetBookingsResp(bookings))
}

type updatePasswordReq struct {
	OldPassword        string `json:"old_password"`
	NewPassword        string `json:"new_password"`
	ConfirmNewPassword string `json:"confirm_new_password"`
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var req updatePasswordReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.authServ.UpdatePassword(r.Context(), mapToUpdatePasswordInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}
