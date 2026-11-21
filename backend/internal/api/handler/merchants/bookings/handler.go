package bookings

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	bookingServ "github.com/miketsu-inc/reservations/backend/internal/service/booking"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	service *bookingServ.Service
}

func NewHandler(s *bookingServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Get("/", h.GetBookings)
	r.Get("/{id}", h.GetByMerchant)
	r.Post("/", h.CreateByMerchant)
	r.Patch("/{id}", h.UpdateByMerchant)
	r.Delete("/{id}", h.CancelByMerchant)

	r.Patch("/{b_id}/participant/{p_id}", h.UpdateParticipantStatus)

	return r
}

type getByMerchantResp struct {
	ID              int                      `json:"id"`
	BookingType     types.BookingType        `json:"booking_type"`
	BookingStatus   types.BookingStatus      `json:"booking_status"`
	FromDate        time.Time                `json:"from_date"`
	ToDate          time.Time                `json:"to_date"`
	IsRecurring     bool                     `json:"is_recurring"`
	MerchantNote    *string                  `json:"merchant_note"`
	EmployeeId      *int                     `json:"employee_id"`
	IsUnassigned    bool                     `json:"is_unassigned"`
	ServiceId       *int                     `json:"service_id"`
	ServiceName     string                   `json:"service_name"`
	ServiceColor    *string                  `json:"service_color"`
	MaxParticipants int                      `json:"max_participants"`
	PricePerPerson  currencyx.FormattedPrice `json:"price_per_person"`
	PriceType       types.PriceType          `json:"price_type"`
	Participants    []bookingParticipantResp `json:"participants"`
	Phases          []bookingPhaseResp       `json:"phases"`
}

type bookingPhaseResp struct {
	Duration  int                    `json:"duration"`
	PhaseType types.ServicePhaseType `json:"phase_type"`
}

type bookingParticipantResp struct {
	Id           int                 `json:"id"`
	CustomerId   uuid.UUID           `json:"customer_id"`
	FirstName    *string             `json:"first_name"`
	LastName     *string             `json:"last_name"`
	CustomerNote *string             `json:"customer_note"`
	Status       types.BookingStatus `json:"status"`
	AvatarUrl    *string             `json:"avatar_url"`
}

func (h *Handler) GetByMerchant(w http.ResponseWriter, r *http.Request) error {
	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid booking id")
	}

	booking, err := h.service.GetByMerchant(r.Context(), urlId)
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "GetByMerchant")
	}

	httputil.Success(w, http.StatusOK, mapToGetByMerchantResp(booking))

	return nil
}

type createByMerchantReq struct {
	Customers    []customerReq     `json:"customers"`
	ServiceId    int               `json:"service_id" validate:"required"`
	EmployeeId   int               `json:"employee_id" validate:"required"`
	TimeStamp    string            `json:"timestamp" validate:"required"`
	MerchantNote *string           `json:"merchant_note"`
	IsRecurring  bool              `json:"is_recurring"`
	Rrule        *recurringRuleReq `json:"recurrence_rule"`
}

type customerReq struct {
	CustomerId  *uuid.UUID `json:"id"`
	FirstName   *string    `json:"first_name"`
	LastName    *string    `json:"last_name"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
}

type recurringRuleReq struct {
	Frequency string   `json:"frequency"`
	Interval  int      `json:"interval"`
	Weekdays  []string `json:"weekdays"`
	UntilDate string   `json:"until_date"`
}

func (h *Handler) CreateByMerchant(w http.ResponseWriter, r *http.Request) error {
	var req createByMerchantReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	input, err := mapToCreateByMerchantInput(req)
	if err != nil {
		return validate.NewError(err.Error())
	}

	err = h.service.CreateByMerchant(r.Context(), input)
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "CreateByMerchant")
	}

	w.WriteHeader(http.StatusCreated)

	return nil
}

// validate:"required" on MerchantNote would fail
// if an empty string arrives as a note
type updateByMerchantReq struct {
	Customers       []customerReq       `json:"customers"`
	TimeStamp       string              `json:"timestamp" validate:"required"`
	MerchantNote    *string             `json:"merchant_note"`
	EmployeeId      int                 `json:"employee_id" validate:"required"`
	BookingStatus   types.BookingStatus `json:"booking_status" validate:"required"`
	UpdateAllFuture bool                `json:"update_all_future"`
}

func (h *Handler) UpdateByMerchant(w http.ResponseWriter, r *http.Request) error {
	var req updateByMerchantReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid booking id")
	}

	input, err := mapToUpdateByMerchantInput(req)
	if err != nil {
		return validate.NewError(err.Error())
	}

	err = h.service.UpdateByMerchant(r.Context(), urlId, input)
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "UpdateByMerchant")
	}

	return nil
}

type cancelByMerchantReq struct {
	CancellationReason string `json:"cancellation_reason"`
	CancelFuture       bool   `json:"cancel_future"`
}

func (h *Handler) CancelByMerchant(w http.ResponseWriter, r *http.Request) error {
	var req cancelByMerchantReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlId, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return validate.NewError("invalid booking id")
	}

	err = h.service.CancelByMerchant(r.Context(), urlId, mapToCancelByMerchantInput(req))
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "CancelByMerchant")
	}

	return nil
}

type updatePaticipantStatusReq struct {
	Status types.BookingStatus `json:"status" validate:"required"`
}

func (h *Handler) UpdateParticipantStatus(w http.ResponseWriter, r *http.Request) error {
	var req updatePaticipantStatusReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	urlBookingId, err := strconv.Atoi(chi.URLParam(r, "b_id"))
	if err != nil {
		return validate.NewError("invalid booking id")
	}

	urlParticipantId, err := strconv.Atoi(chi.URLParam(r, "p_id"))
	if err != nil {
		return validate.NewError("invalid participant id")
	}

	err = h.service.UpdateParticipantStatus(r.Context(), urlBookingId, urlParticipantId, mapToUpdateParticipantStatusInput(req))
	if err != nil {
		return bookingServ.ErrStatus.Resolve(err, "UpdateParticipantStatus")
	}

	return nil
}
