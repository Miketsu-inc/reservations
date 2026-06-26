package merchants

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

type Handler struct {
	service    *merchantServ.Service
	middleware *middleware.Manager
}

func NewHandler(s *merchantServ.Service, m *middleware.Manager) *Handler {
	return &Handler{service: s, middleware: m}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(h.middleware.Language)

		r.Get("/", h.GetInfo)
		r.Get("/services", h.GetServices)
		r.Get("/team", h.GetTeam)

		r.Get("/locations/{locationId}/business-hours/normalized", h.GetNormalizedBusinessHours)

		r.Get("/locations/{locationId}/services/{serviceId}", h.GetServiceDetails)
		r.Get("/locations/{locationId}/summary", h.GetSummary)
		r.Get("/locations/{locationId}/services/{serviceId}/availability", h.GetAvailability)
		r.Get("/locations/{locationId}/services/{serviceId}/availability/next", h.GetNextAvailability)
		r.Get("/locations/{locationId}/services/{serviceId}/availability/disabled-days", h.GetDisabledDays)
	})

	return r
}

type getInfoResp struct {
	Name         string `json:"merchant_name"`
	UrlName      string `json:"url_name"`
	ContactEmail string `json:"contact_email"`
	Introduction string `json:"introduction"`
	Announcement string `json:"announcement"`
	AboutUs      string `json:"about_us"`
	ParkingInfo  string `json:"parking_info"`
	PaymentInfo  string `json:"payment_info"`
	Timezone     string `json:"timezone"`

	LocationId        int            `json:"location_id"`
	Country           *string        `json:"country"`
	City              *string        `json:"city"`
	PostalCode        *string        `json:"postal_code"`
	Address           *string        `json:"address"`
	FormattedLocation string         `json:"formatted_location"`
	GeoPoint          types.GeoPoint `json:"geo_point"`

	BusinessHours           map[int][]timeSlotResp  `json:"business_hours"`
	BusinessHoursStatusResp businessHoursStatusResp `json:"business_hours_status"`
}

type timeSlotResp struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type businessHoursStatusResp struct {
	IsOpen      bool    `json:"is_business_open"`
	CloseTime   *string `json:"close_time"`
	NextOpenDay *int    `json:"next_open_day"`
}

func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
		return
	}

	info, err := h.service.GetInfo(r.Context(), urlName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetInfo(info))
}

type servicesGroupedByCategoryResp struct {
	Id       *int          `json:"id"`
	Name     *string       `json:"name"`
	Sequence *int          `json:"sequence"`
	Services []serviceResp `json:"services"`
}

type serviceResp struct {
	Id              int                       `json:"id"`
	CategoryId      *int                      `json:"category_id"`
	Name            string                    `json:"name"`
	Description     *string                   `json:"description"`
	TotalDuration   int                       `json:"total_duration"`
	Price           *currencyx.FormattedPrice `json:"price"`
	PriceType       types.PriceType           `json:"price_type"`
	MaxParticipants int                       `json:"max_participants"`
	BookingType     types.BookingType         `json:"booking_type"`
	Sequence        int                       `json:"sequence"`
}

func (h *Handler) GetServices(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")
	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
	}

	services, err := h.service.GetServicesGroupedByCategories(r.Context(), urlName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetServicesGroupedByCategories(services))
}

type teamResponse struct {
	Id          int                `json:"id"`
	Role        types.EmployeeRole `json:"role"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	Email       *string            `json:"email" `
	PhoneNumber *string            `json:"phone_number"`
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")
	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
	}

	team, err := h.service.GetTeam(r.Context(), urlName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetTeam(team))
}

func (h *Handler) GetNormalizedBusinessHours(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
		return
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid locationId: %s", err.Error()))
		return
	}

	businessHours, err := h.service.GetNormalizedBusinessHoursPublic(r.Context(), merchantServ.GetNormalizedBusinessHoursPublicInput{
		MerchantUrl: urlName,
		LocationId:  urlLocationId,
	})
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetNormalizedBusinessHoursResp(businessHours))
}

type getServiceDetailsResp struct {
	Id                int                       `json:"id"`
	Name              string                    `json:"name"`
	Description       *string                   `json:"description"`
	TotalDuration     int                       `json:"total_duration"`
	Price             *currencyx.FormattedPrice `json:"price"`
	PriceType         types.PriceType           `json:"price_type"`
	FormattedLocation string                    `json:"formatted_location"`
	GeoPoint          types.GeoPoint            `json:"geo_point"`
	Phases            []phaseResp               `json:"phases"`
}

type phaseResp struct {
	Id        int                    `json:"id"`
	ServiceId int                    `json:"service_id"`
	Name      string                 `json:"name"`
	Sequence  int                    `json:"sequence"`
	Duration  int                    `json:"duration"`
	PhaseType types.ServicePhaseType `json:"phase_type"`
}

func (h *Handler) GetServiceDetails(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
		return
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid serviceId: %s", err.Error()))
		return
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid locationId: %s", err.Error()))
		return
	}

	serviceDetails, err := h.service.GetServiceDetails(r.Context(), urlName, urlServiceId, urlLocationId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetServiceDetailsResp(serviceDetails))
}

type getSummaryResp struct {
	MerchantName      string                    `json:"merchant_name"`
	FormattedLocation string                    `json:"formatted_location"`
	ServiceName       *string                   `json:"service_name"`
	Price             *currencyx.FormattedPrice `json:"price"`
	PriceType         *types.PriceType          `json:"price_type"`
	TotalDuration     *int                      `json:"total_duration"`
	EmployeeFirstName *string                   `json:"employee_first_name"`
	EMployeeLastName  *string                   `json:"employee_last_name"`
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")
	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
		return
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid locationId: %s", err.Error()))
		return
	}

	var urlServiceId *int
	if sId := r.URL.Query().Get("serviceId"); sId != "" {
		parsedId, err := strconv.Atoi(sId)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid serviceId: %w", err))
			return
		}
		urlServiceId = &parsedId
	}

	var urlEmployeeId *int
	if eId := r.URL.Query().Get("employeeId"); eId != "" {
		parsedId, err := strconv.Atoi(eId)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid employeeId %s", err.Error()))
		}
		urlEmployeeId = &parsedId
	}

	summaryInfo, err := h.service.GetSummary(r.Context(), urlName, urlLocationId, urlServiceId, urlEmployeeId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetSummaryResp(summaryInfo))
}

type getAvailabilityResp struct {
	Date        string   `json:"date"`
	IsAvailable bool     `json:"is_available"`
	Morning     []string `json:"morning"`
	Afternoon   []string `json:"afternoon"`
}

func (h *Handler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
		return
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid serviceId: %s", err.Error()))
		return
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid locationId: %s", err.Error()))
		return
	}

	urlStartDate, err := time.Parse(time.RFC3339, r.URL.Query().Get("start"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid date format: %s", err.Error()))
		return
	}

	urlEndDate, err := time.Parse(time.RFC3339, r.URL.Query().Get("end"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid date format: %s", err.Error()))
		return
	}

	availability, err := h.service.GetAvailability(r.Context(), urlName, urlServiceId, urlLocationId, urlStartDate, urlEndDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetAvailabilityResp(availability))
}

type getNextAvailabilityResp struct {
	FromDate            *time.Time `json:"from_date"`
	ToDate              *time.Time `json:"to_date"`
	CurrentParticipants *int       `json:"current_participants"`
	Employee            *int       `json:"employee"`
}

func (h *Handler) GetNextAvailability(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
		return
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid serviceId: %s", err.Error()))
		return
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid locationId: %s", err.Error()))
		return
	}

	nextAvailability, err := h.service.GetNextAvailability(r.Context(), urlName, urlServiceId, urlLocationId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetNextAvailabilityResp(nextAvailability))
}

type getDisabledDaysResp struct {
	ClosedDays []int     `json:"closed_days"`
	MinDate    time.Time `json:"min_date"`
	MaxDate    time.Time `json:"max_date"`
}

func (h *Handler) GetDisabledDays(w http.ResponseWriter, r *http.Request) {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name"))
		return
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid serviceId: %s", err.Error()))
		return
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid locationId: %s", err.Error()))
		return
	}

	disabledDays, err := h.service.GetDisabledDays(r.Context(), urlName, urlServiceId, urlLocationId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	httputil.Success(w, http.StatusOK, mapToGetDisabledDaysResp(disabledDays))
}
