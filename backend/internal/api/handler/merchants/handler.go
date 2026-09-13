package merchants

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	externalcalendarServ "github.com/miketsu-inc/reservations/backend/internal/service/externalcalendar"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

// merchant routes are in router.go due to them being at the
// same level as the other authenticated merchant routes
// but not being grouped under a subroute
type Handler struct {
	service         *merchantServ.Service
	extcalendarServ *externalcalendarServ.Service
}

func NewHandler(s *merchantServ.Service, extcalendarServ *externalcalendarServ.Service) *Handler {
	return &Handler{service: s, extcalendarServ: extcalendarServ}
}

type meResp struct {
	UserId     uuid.UUID          `json:"user_id"`
	MerchantId uuid.UUID          `json:"merchant_id"`
	LocationId int                `json:"location_id"`
	EmployeeId int                `json:"employee_id"`
	Role       types.EmployeeRole `json:"role"`
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) error {
	actor := actor.MustGetFromContext(r.Context())

	httputil.Success(w, http.StatusOK, mapToMeResp(actor))

	return nil
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	err := h.service.Delete(r.Context())
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "Delete")
	}

	return nil
}

type updateNameReq struct {
	Name string `json:"name" validate:"required"`
}

func (h *Handler) UpdateName(w http.ResponseWriter, r *http.Request) error {
	var req updateNameReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	err := h.service.UpdateName(r.Context(), mapToUpdateNameInput(req))
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "UpdateName")
	}

	return nil
}

type getDashboardResp struct {
	PeriodStart      time.Time               `json:"period_start"`
	PeriodEnd        time.Time               `json:"period_end"`
	UpcomingBookings []bookingDetailsResp    `json:"upcoming_bookings"`
	LatestBookings   []bookingDetailsResp    `json:"latest_bookings"`
	LowStockProducts []lowStockProductResp   `json:"low_stock_products"`
	Statistics       dashboardStatisticsResp `json:"statistics"`
}

type bookingDetailsResp struct {
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

type lowStockProductResp struct {
	Id            int     `json:"id"`
	Name          string  `json:"name"`
	MaxAmount     int     `json:"max_amount"`
	CurrentAmount int     `json:"current_amount"`
	Unit          string  `json:"unit"`
	FillRatio     float64 `json:"fill_ratio"`
}

type dashboardStatisticsResp struct {
	Revenue               []revenueStatResp `json:"revenue"`
	RevenueSum            string            `json:"revenue_sum"`
	RevenueChange         int               `json:"revenue_change"`
	Bookings              int               `json:"bookings"`
	BookingsChange        int               `json:"bookings_change"`
	Cancellations         int               `json:"cancellations"`
	CancellationsChange   int               `json:"cancellations_change"`
	AverageDuration       int               `json:"average_duration"`
	AverageDurationChange int               `json:"average_duration_change"`
}

// TODO: value is of numeric type so float might not be the best
// type to return here
type revenueStatResp struct {
	Value float64   `json:"value"`
	Day   time.Time `json:"day"`
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) error {
	urlDate, err := time.Parse(time.RFC3339, r.URL.Query().Get("date"))
	if err != nil {
		return validate.NewError(fmt.Sprintf("invalid date: %s", err.Error()))
	}

	urlPeriod, err := strconv.Atoi(r.URL.Query().Get("period"))
	if err != nil {
		return validate.NewError("invalid period")
	}

	dashboard, err := h.service.GetDashboard(r.Context(), urlDate, urlPeriod)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetDashboard")
	}

	httputil.Success(w, http.StatusOK, mapToGetDashboardResp(dashboard))

	return nil
}

type checkUrlReq struct {
	Name string `json:"merchant_name" validate:"required"`
}

type checkUrlResp struct {
	Name string `json:"merchant_url"`
}

func (h *Handler) CheckUrl(w http.ResponseWriter, r *http.Request) error {
	var req checkUrlReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	merchantUrl, err := h.service.CheckUrl(r.Context(), mapToCheckUrlInput(req))
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "CheckUrl")
	}

	httputil.Success(w, http.StatusOK, mapToCheckUrlResp(merchantUrl))

	return nil
}

type getSettingsResp struct {
	Name             string                 `json:"merchant_name"`
	ContactEmail     string                 `json:"contact_email"`
	Introduction     string                 `json:"introduction"`
	Announcement     string                 `json:"announcement"`
	AboutUs          string                 `json:"about_us"`
	ParkingInfo      string                 `json:"parking_info"`
	PaymentInfo      string                 `json:"payment_info"`
	CancelDeadline   int                    `json:"cancel_deadline"`
	BookingWindowMin int                    `json:"booking_window_min"`
	BookingWindowMax int                    `json:"booking_window_max"`
	BufferTime       int                    `json:"buffer_time"`
	ApprovalPolicy   types.ApprovalType     `json:"approval_policy"`
	Timezone         string                 `json:"timezone"`
	BusinessHours    map[int][]timeSlotResp `json:"business_hours"`

	LocationId        int     `json:"location_id"`
	Country           *string `json:"country"`
	City              *string `json:"city"`
	PostalCode        *string `json:"postal_code"`
	Address           *string `json:"address"`
	FormattedLocation string  `json:"formatted_location"`
}

type timeSlotResp struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) error {
	settings, err := h.service.GetSettings(r.Context())
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetSettings")
	}

	httputil.Success(w, http.StatusOK, mapToGetSettingsResp(settings))

	return nil
}

type updateSettingsReq struct {
	Introduction     string                 `json:"introduction"`
	Announcement     string                 `json:"announcement"`
	AboutUs          string                 `json:"about_us"`
	ParkingInfo      string                 `json:"parking_info"`
	PaymentInfo      string                 `json:"payment_info"`
	CancelDeadline   int                    `json:"cancel_deadline"`
	BookingWindowMin int                    `json:"booking_window_min"`
	BookingWindowMax int                    `json:"booking_window_max"`
	BufferTime       int                    `json:"buffer_time"`
	ApprovalPolicy   types.ApprovalType     `json:"approval_policy" validate:"required"`
	BusinessHours    map[int][]timeSlotResp `json:"business_hours"`
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) error {
	var req updateSettingsReq

	if err := validate.ParseStruct(r, &req); err != nil {
		return err
	}

	updateSettingsInput, err := mapToUpdateSettingsInput(req)
	if err != nil {
		return validate.NewError(err.Error())
	}

	err = h.service.UpdateSettings(r.Context(), updateSettingsInput)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "UpdateSettings")
	}

	return nil
}

func (h *Handler) GetNormalizedBusinessHours(w http.ResponseWriter, r *http.Request) error {
	businessHours, err := h.service.GetNormalizedBusinessHours(r.Context())
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetNormalizedBusinessHours")
	}

	httputil.Success(w, http.StatusOK, mapToGetNormalizedBusinessHoursResp(businessHours))

	return nil
}

type getTeamMembersForCalendarResp struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (h *Handler) GetTeamForCalendar(w http.ResponseWriter, r *http.Request) error {
	// TODO: should be in team service
	teamMembers, err := h.service.GetTeamForCalendar(r.Context())
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetTeamForCalendar")
	}

	httputil.Success(w, http.StatusOK, mapToGetTeamMembersForCalendarResp(teamMembers))

	return nil
}

type getServicesForCalendarResp struct {
	Id       *int                  `json:"id"`
	Name     *string               `json:"name"`
	Services []calendarServiceResp `json:"services"`
}

type calendarServiceResp struct {
	Id              int                       `json:"id"`
	Name            string                    `json:"name"`
	Duration        int                       `json:"duration"`
	Price           *currencyx.FormattedPrice `json:"price"`
	PriceType       types.PriceType           `json:"price_type"`
	Color           string                    `json:"color"`
	BookingType     types.BookingType         `json:"booking_type"`
	MaxParticipants int                       `json:"max_participants"`
}

func (h *Handler) GetServicesForCalendar(w http.ResponseWriter, r *http.Request) error {
	// TODO: should be in catalog service
	services, err := h.service.GetServicesForCalendar(r.Context())
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetServicesForCalendar")
	}

	httputil.Success(w, http.StatusOK, mapToGetServicesForCalendarResp(services))

	return nil
}

type getCustomersForCalendarResp struct {
	CustomerId  uuid.UUID  `json:"customer_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
	BirthDay    *time.Time `json:"birthday"`
	IsDummy     bool       `json:"is_dummy"`
	LastVisited *time.Time `json:"last_visited"`
}

func (h *Handler) GetCustomersForCalendar(w http.ResponseWriter, r *http.Request) error {
	// TODO: should be in customer service
	customers, err := h.service.GetCustomersForCalendar(r.Context())
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetCustomersForCalendar")
	}

	httputil.Success(w, http.StatusOK, mapToGetCustomersForCalendarResp(customers))

	return nil
}

type getCalendarEventsResp struct {
	Bookings     []bookingForCalendar `json:"bookings"`
	BlockedTimes []blockedTime        `json:"blocked_times"`
}

type bookingForCalendar struct {
	ID              int                             `json:"id"`
	BookingType     types.BookingType               `json:"booking_type"`
	BookingStatus   types.BookingStatus             `json:"booking_status"`
	FromDate        time.Time                       `json:"from_date"`
	ToDate          time.Time                       `json:"to_date"`
	IsRecurring     bool                            `json:"is_recurring"`
	Duration        int                             `json:"duration"`
	MerchantNote    *string                         `json:"merchant_note"`
	EmployeeId      *int                            `json:"employee_id"`
	ServiceId       *int                            `json:"service_id"`
	ServiceName     string                          `json:"service_name"`
	ServiceColor    *string                         `json:"service_color"`
	MaxParticipants int                             `json:"max_participants"`
	Price           currencyx.FormattedPrice        `json:"price"`
	PriceType       types.PriceType                 `json:"price_type"`
	Participants    []bookingParticipantForCalendar `json:"participants"`
}

type bookingParticipantForCalendar struct {
	Id           int                 `json:"id"`
	CustomerId   uuid.UUID           `json:"customer_id"`
	FirstName    *string             `json:"first_name"`
	LastName     *string             `json:"last_name"`
	CustomerNote *string             `json:"customer_note"`
	Status       types.BookingStatus `json:"status"`
}

type blockedTime struct {
	ID            int       `json:"id"`
	EmployeeIds   []int     `json:"employee_ids"`
	Name          string    `json:"name"`
	FromDate      time.Time `json:"from_date"`
	ToDate        time.Time `json:"to_date"`
	AllDay        bool      `json:"all_day"`
	Icon          *string   `json:"icon"`
	BlockedTypeId *int      `json:"blocked_type_id"`
}

func (h *Handler) GetCalendarEvents(w http.ResponseWriter, r *http.Request) error {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	bookings, err := h.service.GetCalendarEvents(r.Context(), start, end)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetCalendarEvents")
	}

	httputil.Success(w, http.StatusOK, mapToGetCalendarEventsResp(bookings))

	return nil
}

func (h *Handler) GoogleCalendar(w http.ResponseWriter, r *http.Request) error {
	url, err := h.extcalendarServ.GoogleCalendar(r.Context())
	if err != nil {
		return externalcalendarServ.ErrStatus.Resolve(err, "GoogleCalendar")
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)

	return nil
}
