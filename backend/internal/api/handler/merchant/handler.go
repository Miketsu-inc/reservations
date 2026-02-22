package merchant

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
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
	service *merchantServ.Service
}

func NewHandler(s *merchantServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.service.Delete(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, err)
		return
	}
}

type updateNameReq struct {
	Name string `json:"name" validate:"required"`
}

func (h *Handler) UpdateName(w http.ResponseWriter, r *http.Request) {
	var req updateNameReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
	}

	err := h.service.UpdateName(r.Context(), mapToUpdateNameInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadGateway, err)
		return
	}
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
	ID              int                      `json:"id"`
	FromDate        time.Time                `json:"from_date"`
	ToDate          time.Time                `json:"to_date"`
	CustomerNote    *string                  `json:"customer_note"`
	MerchantNote    *string                  `json:"merchant_note"`
	ServiceName     string                   `json:"service_name"`
	ServiceColor    string                   `json:"service_color"`
	ServiceDuration int                      `json:"service_duration"`
	Price           currencyx.FormattedPrice `json:"price"`
	Cost            currencyx.FormattedPrice `json:"cost"`
	FirstName       *string                  `json:"first_name"`
	LastName        *string                  `json:"last_name"`
	PhoneNumber     *string                  `json:"phone_number"`
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

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	urlDate, err := time.Parse(time.RFC3339, r.URL.Query().Get("date"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid date: %s", err.Error()))
		return
	}

	urlPeriod, err := strconv.Atoi(r.URL.Query().Get("period"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid period: %s", err.Error()))
		return
	}

	dashboard, err := h.service.GetDashboard(r.Context(), urlDate, urlPeriod)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetDashboardResp(dashboard))
}

type checkUrlReq struct {
	Name string `json:"merchant_name"`
}

type checkUrlResp struct {
	Name string `json:"merchant_name"`
}

func (h *Handler) CheckUrl(w http.ResponseWriter, r *http.Request) {
	var req checkUrlReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	merchantUrl, err := h.service.CheckUrl(r.Context(), mapToCheckUrlInput(req))
	if err != nil {
		switch e := err.(type) {
		case merchantServ.ErrMerchantUrlNotUnique:
			httputil.WriteJSON(w, http.StatusConflict, map[string]map[string]string{
				"error": {
					"message":      err.Error(),
					"merchant_url": e.URL},
			})
			return

		default:
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}
	}

	httputil.Success(w, http.StatusOK, mapToCheckUrlResp(merchantUrl))

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

func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.GetSettings(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetSettingsResp(settings))
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
	BusinessHours    map[int][]timeSlotResp `json:"business_hours"`
}

func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req updateSettingsReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.UpdateSettings(r.Context(), mapToUpdateSettingsInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

func (h *Handler) GetNormalizedBusinessHours(w http.ResponseWriter, r *http.Request) {
	businessHours, err := h.service.GetNormalizedBusinessHours(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetNormalizedBusinessHoursResp(businessHours))
}

type getPreferencesResp struct {
	FirstDayOfWeek     string `json:"first_day_of_week"`
	TimeFormat         string `json:"time_format"`
	CalendarView       string `json:"calendar_view"`
	CalendarViewMobile string `json:"calendar_view_mobile"`
	StartHour          string `json:"start_hour"`
	EndHour            string `json:"end_hour"`
	TimeFrequency      string `json:"time_frequency"`
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	preferences, err := h.service.GetPreferences(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetPreferencesResp(preferences))
}

type updatePreferencesReq struct {
	FirstDayOfWeek     string `json:"first_day_of_week"`
	TimeFormat         string `json:"time_format"`
	CalendarView       string `json:"calendar_view"`
	CalendarViewMobile string `json:"calendar_view_mobile"`
	StartHour          string `json:"start_hour"`
	EndHour            string `json:"end_hour"`
	TimeFrequency      string `json:"time_frequency"`
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	var req updatePreferencesReq

	if err := validate.ParseStruct(r, &req); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.UpdatePreferences(r.Context(), mapToUpdatePreferencesInput(req))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}
}

type getTeamMembersForCalendarResp struct {
	Id        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (h *Handler) GetTeamMembersForCalendar(w http.ResponseWriter, r *http.Request) {
	teamMembers, err := h.service.GetTeamMembersForCalendar(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetTeamMembersForCalendarResp(teamMembers))
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

func (h *Handler) GetServicesForCalendar(w http.ResponseWriter, r *http.Request) {
	services, err := h.service.GetServicesForCalendar(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetServicesForCalendarResp(services))
}

type getCustomersForCalendarResp struct {
	Id          uuid.UUID  `json:"id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phone_number"`
	BirthDay    *time.Time `json:"birthday"`
	IsDummy     bool       `json:"is_dummy"`
	LastVisited *time.Time `json:"last_visited"`
}

func (h *Handler) GetCustomersForCalendar(w http.ResponseWriter, r *http.Request) {
	customers, err := h.service.GetCustomersForCalendar(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetCustomersForCalendarResp(customers))
}
