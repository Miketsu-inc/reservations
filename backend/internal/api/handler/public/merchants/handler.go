package merchants

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware"
	catalogServ "github.com/miketsu-inc/reservations/backend/internal/service/catalog"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	teamServ "github.com/miketsu-inc/reservations/backend/internal/service/team"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Handler struct {
	merchantServ *merchantServ.Service
	catalogServ  *catalogServ.Service
	teamServ     *teamServ.Service
	middleware   *middleware.Manager
}

func NewHandler(
	merchant *merchantServ.Service,
	catalog *catalogServ.Service,
	team *teamServ.Service,
	m *middleware.Manager,
) *Handler {
	return &Handler{
		merchantServ: merchant,
		catalogServ:  catalog,
		teamServ:     team,
		middleware:   m,
	}
}

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Group(func(r *httputil.Router) {
		r.UseFunc(h.middleware.Language)

		r.Get("/", h.GetInfo)
		r.Get("/services", h.GetServices)
		r.Get("/team", h.GetTeam)

		r.Get("/locations/{locationId}/business-hours/normalized", h.GetNormalizedBusinessHours)

		r.Get("/locations/{locationId}/services/{serviceId}", h.GetServiceDetails)
		r.Get("/locations/{locationId}/summary", h.GetSummary)
		r.Get("/locations/{locationId}/services/{serviceId}/availability/available-days", h.GetDayAvailability)
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

func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	info, err := h.merchantServ.GetInfo(r.Context(), urlName)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetInfo")
	}

	httputil.Success(w, http.StatusOK, mapToGetInfo(info))

	return nil
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

func (h *Handler) GetServices(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")
	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	services, err := h.catalogServ.GetServicesGroupedByCategories(r.Context(), urlName)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetServicesGroupedByCategories")
	}

	httputil.Success(w, http.StatusOK, mapToGetServicesGroupedByCategories(services))

	return nil
}

type teamResponse struct {
	Id          int                `json:"id"`
	Role        types.EmployeeRole `json:"role"`
	FirstName   string             `json:"first_name"`
	LastName    string             `json:"last_name"`
	Email       *string            `json:"email" `
	PhoneNumber *string            `json:"phone_number"`
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")
	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	team, err := h.teamServ.GetTeamByMerchantName(r.Context(), urlName)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetTeam")
	}

	httputil.Success(w, http.StatusOK, mapToGetTeam(team))

	return nil
}

func (h *Handler) GetNormalizedBusinessHours(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	businessHours, err := h.merchantServ.GetNormalizedBusinessHoursPublic(r.Context(), merchantServ.GetNormalizedBusinessHoursPublicInput{
		MerchantUrl: urlName,
		LocationId:  urlLocationId,
	})
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetNormalizedBusinessHoursPublic")
	}

	httputil.Success(w, http.StatusOK, mapToGetNormalizedBusinessHoursResp(businessHours))

	return nil
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

func (h *Handler) GetServiceDetails(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	serviceDetails, err := h.catalogServ.GetServiceDetails(r.Context(), urlName, urlServiceId, urlLocationId)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetServiceDetails")
	}

	httputil.Success(w, http.StatusOK, mapToGetServiceDetailsResp(serviceDetails))

	return nil
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

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")
	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	var urlServiceId *int
	if sId := r.URL.Query().Get("serviceId"); sId != "" {
		parsedId, err := strconv.Atoi(sId)
		if err != nil {
			return validate.NewError("invalid service id")
		}
		urlServiceId = &parsedId
	}

	var urlEmployeeId *int
	if eId := r.URL.Query().Get("employeeId"); eId != "" {
		parsedId, err := strconv.Atoi(eId)
		if err != nil {
			return validate.NewError("invalid employee id")
		}
		urlEmployeeId = &parsedId
	}

	summaryInfo, err := h.merchantServ.GetSummary(r.Context(), urlName, urlLocationId, urlServiceId, urlEmployeeId)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetSummary")
	}

	httputil.Success(w, http.StatusOK, mapToGetSummaryResp(summaryInfo))

	return nil
}

type getDayAvailabilityResp struct {
	Date        string `json:"date"`
	IsAvailable bool   `json:"is_available"`
}

func (h *Handler) GetDayAvailability(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	var employeeId *int

	if empIdStr := r.URL.Query().Get("employee_id"); empIdStr != "" {
		parsedID, err := strconv.Atoi(empIdStr)
		if err != nil {
			return validate.NewError("invalid employee id")
		}
		employeeId = &parsedID
	}

	availability, err := h.merchantServ.GetDayAvailability(r.Context(), urlName, urlServiceId, urlLocationId, employeeId)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetDayAvailability")
	}

	httputil.Success(w, http.StatusOK, mapToGetDayAvailabilityResp(availability))

	return nil
}

type getAvailabilityResp struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
}

func (h *Handler) GetAvailability(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	urlBookingDay, err := time.Parse(time.DateOnly, r.URL.Query().Get("date"))
	if err != nil {
		return validate.NewError("invalid date format")
	}

	var employeeId *int

	if empIdStr := r.URL.Query().Get("employee_id"); empIdStr != "" {
		parsedID, err := strconv.Atoi(empIdStr)
		if err != nil {
			return validate.NewError("invalid employee id")
		}
		employeeId = &parsedID
	}

	availability, err := h.merchantServ.GetAvailabilityForDay(r.Context(), urlName, urlServiceId, urlLocationId, employeeId, urlBookingDay)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetAvailability")
	}

	httputil.Success(w, http.StatusOK, mapToGetAvailabilityResp(availability))

	return nil
}

type getNextAvailabilityResp struct {
	FromDate            *time.Time `json:"from_date"`
	ToDate              *time.Time `json:"to_date"`
	CurrentParticipants *int       `json:"current_participants"`
	Employee            *int       `json:"employee"`
}

func (h *Handler) GetNextAvailability(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	nextAvailability, err := h.merchantServ.GetNextAvailability(r.Context(), urlName, urlServiceId, urlLocationId)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetNextAvailability")
	}

	httputil.Success(w, http.StatusOK, mapToGetNextAvailabilityResp(nextAvailability))

	return nil
}

type getDisabledDaysResp struct {
	ClosedDays []int     `json:"closed_days"`
	MinDate    time.Time `json:"min_date"`
	MaxDate    time.Time `json:"max_date"`
}

func (h *Handler) GetDisabledDays(w http.ResponseWriter, r *http.Request) error {
	urlName := chi.URLParam(r, "merchantName")

	if urlName == "" {
		return validate.NewError("invalid merchant name")
	}

	urlServiceId, err := strconv.Atoi(chi.URLParam(r, "serviceId"))
	if err != nil {
		return validate.NewError("invalid service id")
	}

	urlLocationId, err := strconv.Atoi(chi.URLParam(r, "locationId"))
	if err != nil {
		return validate.NewError("invalid location id")
	}

	disabledDays, err := h.merchantServ.GetDisabledDays(r.Context(), urlName, urlServiceId, urlLocationId)
	if err != nil {
		return merchantServ.ErrStatus.Resolve(err, "GetDisabledDays")
	}

	httputil.Success(w, http.StatusOK, mapToGetDisabledDaysResp(disabledDays))

	return nil
}
