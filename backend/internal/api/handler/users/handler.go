package users

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	authServ "github.com/miketsu-inc/reservations/backend/internal/service/auth"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	mediaServ "github.com/miketsu-inc/reservations/backend/internal/service/media"
	userServ "github.com/miketsu-inc/reservations/backend/internal/service/user"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service     *userServ.Service
	bookingServ *bookingServ.Service
	authServ    *authServ.Service
	mediaServ   *mediaServ.Service
	middleware  *middleware.Manager
}

func NewHandler(s *userServ.Service, b *bookingServ.Service, a *authServ.Service, ms *mediaServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, bookingServ: b, authServ: a, mediaServ: ms, middleware: m}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.JwtAuthentication)

		r.Put("/", h.Edit)
		r.Delete("/", h.Delete)

		r.Get("/bookings", h.GetBookings)
		r.Get("/bookings/counts", h.GetBookingCounts)
		r.Get("/bookings/{bookingId}", h.GetBookingById)
		r.Delete("/bookings/{bookingId}", h.CancelBooking)

		r.Put("/password", h.UpdatePassword)

		r.WithFunc(h.middleware.RateLimitByUser(30, time.Hour)).Post("/avatar", h.CreateAvatarUpload)
		r.Delete("/avatar", h.RemoveAvatar)
	})

	return r
}

type editReq struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required,e164"`
	Email       string `json:"email" validate:"required,email"`
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) error {
	var req editReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.Edit(r.Context(), mapToEditInput(req))
	if err != nil {
		return userServ.ErrStatus.Resolve(err, "Edit")
	}

	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	err := h.service.Delete(r.Context())
	if err != nil {
		return userServ.ErrStatus.Resolve(err, "Delete")
	}

	jwt.DeleteJwts(w)

	return nil
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
	MerchantLogoUrl   *string                  `json:"merchant_logo_url"`
	FormattedLocation string                   `json:"formatted_location"`
	ServiceName       string                   `json:"service_name"`
	EmployeeFirstName *string                  `json:"employee_first_name"`
	EmployeeLastName  *string                  `json:"employee_last_name"`
}

func (h *Handler) GetBookings(w http.ResponseWriter, r *http.Request) error {
	urlStatus := r.URL.Query().Get("status")
	if urlStatus != "upcoming" && urlStatus != "completed" && urlStatus != "cancelled" {
		return validate.NewError("invalid status query parameter")
	}

	urlCursor := r.URL.Query().Get("cursor")

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		return validate.NewError("invalid limit query parameter")
	}

	if limit < 1 || limit > 10 {
		return validate.NewError("limit must be between 1 and 10")
	}

	bookings, err := h.bookingServ.GetForUser(r.Context(), urlStatus, urlCursor, limit)
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "GetForUser")
	}

	httputil.Success(w, http.StatusOK, mapToGetBookingsResp(bookings))

	return nil
}

type getBookingCountsResp struct {
	Upcoming  int `json:"upcoming"`
	Completed int `json:"completed"`
	Cancelled int `json:"cancelled"`
}

func (h *Handler) GetBookingCounts(w http.ResponseWriter, r *http.Request) error {
	counts, err := h.bookingServ.GetCountsForUser(r.Context())
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "GetCountsForUser")
	}

	httputil.Success(w, http.StatusOK, mapToGetBookingCountsResp(counts))
	return nil
}

type bookingDetailsResp struct {
	Id                int                      `json:"id"`
	Status            string                   `json:"status"`
	ParticipantStatus string                   `json:"participant_status"`
	BookingType       string                   `json:"booking_type"`
	IsRecurring       bool                     `json:"is_recurring"`
	FromDate          time.Time                `json:"from_date"`
	ToDate            time.Time                `json:"to_date"`
	Price             currencyx.FormattedPrice `json:"price"`
	PriceType         string                   `json:"price_type"`
	ServiceName       string                   `json:"service_name"`
	MerchantName      string                   `json:"merchant_name"`
	MerchantUrl       string                   `json:"merchant_url"`
	CancelDeadline    int                      `json:"cancel_deadline"`
	FormattedLocation string                   `json:"formatted_location"`
	Longitude         float64                  `json:"longitude"`
	Latitude          float64                  `json:"latitude"`
	EmployeeFirstName *string                  `json:"employee_first_name"`
	EmployeeLastName  *string                  `json:"employee_last_name"`
	EmployeeAvatarUrl *string                  `json:"employee_avatar_url"`
	Images            []imageResp              `json:"images"`
}

type imageResp struct {
	Url          string `json:"url"`
	ThumbnailUrl string `json:"thumbnail_url"`
	AltText      string `json:"alt_text"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

func (h *Handler) GetBookingById(w http.ResponseWriter, r *http.Request) error {
	bookingId, err := strconv.Atoi(chi.URLParam(r, "bookingId"))
	if err != nil {
		return validate.NewError("invalid booking id")
	}

	booking, err := h.bookingServ.GetForUserById(r.Context(), bookingId)
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "GetForUserById")
	}

	httputil.Success(w, http.StatusOK, mapToGetBookingDetailsResp(booking))

	return nil
}

func (h *Handler) CancelBooking(w http.ResponseWriter, r *http.Request) error {
	bookingId, err := strconv.Atoi(chi.URLParam(r, "bookingId"))
	if err != nil {
		return validate.NewError("invalid booking id")
	}

	err = h.bookingServ.CancelByCustomer(r.Context(), bookingId)
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "CancelBooking")
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

type updatePasswordReq struct {
	OldPassword string `json:"old_password" validate:"required,ascii"`
	NewPassword string `json:"new_password" validate:"required,ascii"`
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) error {
	var req updatePasswordReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	tokens, err := h.authServ.UpdatePassword(r.Context(), mapToUpdatePasswordInput(req))
	if err != nil {
		return authServ.ErrStatus.Resolve(err, "UpdatePassword")
	}

	jwt.SetJwtCookie(w, jwt.AccessToken, tokens.AccessToken)
	jwt.SetJwtCookie(w, jwt.RefreshToken, tokens.RefreshToken)

	return nil
}

type createUploadReq struct {
	ContentType string `json:"content_type" validate:"required,eq=image/jpeg|eq=image/png|eq=image/webp"`
	Size        int64  `json:"size" validate:"required,min=1"`
}

type uploadResp struct {
	ImageId   uuid.UUID         `json:"image_id"`
	UploadUrl string            `json:"upload_url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time         `json:"expires_at"`
}

func (h *Handler) CreateAvatarUpload(w http.ResponseWriter, r *http.Request) error {
	var req createUploadReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	upload, err := h.mediaServ.CreateUserAvatarUpload(r.Context(), mapToUploadInput(req))
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "CreateAvatarUpload")
	}

	httputil.Success(w, http.StatusCreated, mapToUploadResp(upload))

	return nil
}

func (h *Handler) RemoveAvatar(w http.ResponseWriter, r *http.Request) error {
	err := h.mediaServ.RemoveUserAvatar(r.Context())
	if err != nil {
		return mediaServ.ErrStatus.Resolve(err, "RemoveAvatar")
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
