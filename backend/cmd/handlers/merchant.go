package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bojanz/currency"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/lang"
	"github.com/miketsu-inc/reservations/backend/cmd/types/booking"
	"github.com/miketsu-inc/reservations/backend/cmd/types/location"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/email"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
	"github.com/teambition/rrule-go"
)

type Merchant struct {
	Postgresdb database.PostgreSQL
}

type FormattedAvailableTimes struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
}

func CalculateAvailableTimes(reserved []database.BookingTime, blockedTimes []database.BlockedTimes, servicePhases []database.PublicServicePhase, serviceDuration int, BufferTime int,
	BookingWindowMin int, bookingDay time.Time, businessHours []database.TimeSlot, currentTime time.Time, merchantTz *time.Location) FormattedAvailableTimes {

	year, month, day := bookingDay.Date()
	totalDuration := time.Duration(serviceDuration) * time.Minute
	bufferDuration := time.Duration(BufferTime) * time.Minute
	bookingDeadlineDuration := time.Duration(BookingWindowMin) * time.Minute

	morning := []string{}
	afternoon := []string{}

	for _, blocked := range blockedTimes {
		if blocked.AllDay {
			return FormattedAvailableTimes{
				Morning:   morning,
				Afternoon: afternoon,
			}
		}
	}

	now := currentTime.In(merchantTz)

	stepSize := 15 * time.Minute

	for _, slot := range businessHours {
		startTime, _ := time.Parse("15:04:05", slot.StartTime)
		endTime, _ := time.Parse("15:04:05", slot.EndTime)

		// buisness hours are NOT an absolute point in time,
		// their timezone should be in the same timzone as the merchant is in
		// for golang before/after to work correctly
		businessStart := time.Date(year, month, day, startTime.Hour(), startTime.Minute(), 0, 0, merchantTz)
		businessEnd := time.Date(year, month, day, endTime.Hour(), endTime.Minute(), 0, 0, merchantTz)

		bookingStart := businessStart

		for bookingStart.Add(totalDuration).Before(businessEnd) || bookingStart.Add(totalDuration).Equal(businessEnd) {
			if bookingStart.Before(now.Add(bookingDeadlineDuration)) {
				bookingStart = bookingStart.Add(stepSize)
				continue
			}

			available := true

			phaseStart := bookingStart
			for _, phase := range servicePhases {
				phaseDuration := time.Duration(phase.Duration) * time.Minute
				phaseEnd := phaseStart.Add(phaseDuration)

				if phase.PhaseType == "active" {

					for _, blocked := range blockedTimes {
						if !blocked.AllDay {
							blockedFrom := blocked.FromDate.In(merchantTz)
							blockedTo := blocked.ToDate.In(merchantTz)

							if phaseStart.Before(blockedTo) && phaseEnd.After(blockedFrom) {
								bookingStart = bookingStart.Add(stepSize)

								available = false
								break
							}
						}
					}

					if !available {
						break
					}

					for _, booking := range reserved {
						reservedFromDate := booking.From_date.In(merchantTz).Add(-bufferDuration)
						reservedToDate := booking.To_date.In(merchantTz).Add(bufferDuration)

						if phaseStart.Before(reservedToDate) && phaseEnd.After(reservedFromDate) {
							bookingStart = bookingStart.Add(stepSize)

							available = false
							break
						}
					}
				}

				if !available {
					break
				}

				phaseStart = phaseEnd
			}

			if available {
				formattedTime := fmt.Sprintf("%02d:%02d", bookingStart.Hour(), bookingStart.Minute())

				if bookingStart.Hour() < 12 {
					morning = append(morning, formattedTime)
				} else if bookingStart.Hour() >= 12 {
					afternoon = append(afternoon, formattedTime)
				}

				bookingStart = bookingStart.Add(stepSize)
			}
		}
	}

	availableTimes := FormattedAvailableTimes{
		Morning:   morning,
		Afternoon: afternoon,
	}

	return availableTimes
}

type MultiDayAvailableTimes struct {
	Date        string   `json:"date"`
	IsAvailable bool     `json:"is_available"`
	Morning     []string `json:"morning"`
	Afternoon   []string `json:"afternoon"`
}

func CalculateAvailableTimesPeriod(reservedForPeriod []database.BookingTime, blockedTimes []database.BlockedTimes, servicePhases []database.PublicServicePhase, serviceDuration int, bufferTime int, bookingindowMin int,
	startDate time.Time, endDate time.Time, businessHours map[int][]database.TimeSlot, currentTime time.Time, merchantTz *time.Location) []MultiDayAvailableTimes {

	results := []MultiDayAvailableTimes{}

	reservationsByDate := make(map[string][]database.BookingTime)
	for _, booking := range reservedForPeriod {
		date := booking.From_date.In(merchantTz).Format("2006-01-02")
		reservationsByDate[date] = append(reservationsByDate[date], booking)
	}

	for d := startDate.In(merchantTz); !d.After(endDate.In(merchantTz)); d = d.AddDate(0, 0, 1) {
		businessHoursForDay := businessHours[int(d.Weekday())]
		if len(businessHoursForDay) == 0 {
			continue
		}

		day := d.Format("2006-01-02")
		reservedForDay := reservationsByDate[day]

		blockedForDay := filterBlockedTimesForDay(blockedTimes, d, merchantTz)

		dayResult := CalculateAvailableTimes(reservedForDay, blockedForDay, servicePhases, serviceDuration, bufferTime, bookingindowMin, d, businessHoursForDay, currentTime, merchantTz)

		isAvailable := len(dayResult.Morning) > 0 || len(dayResult.Afternoon) > 0

		results = append(results, MultiDayAvailableTimes{
			Date:        d.Format("2006-01-02"),
			IsAvailable: isAvailable,
			Morning:     dayResult.Morning,
			Afternoon:   dayResult.Afternoon,
		})
	}

	return results
}

func filterBlockedTimesForDay(blockedTimes []database.BlockedTimes, day time.Time, tz *time.Location) []database.BlockedTimes {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, tz)
	dayEnd := dayStart.AddDate(0, 0, 1)

	filtered := []database.BlockedTimes{}
	for _, blocked := range blockedTimes {
		blockedFrom := blocked.FromDate.In(tz)
		blockedTo := blocked.ToDate.In(tz)
		if blockedFrom.Before(dayEnd) && blockedTo.After(dayStart) {
			filtered = append(filtered, blocked)
		}
	}

	return filtered
}

func (m *Merchant) InfoByName(w http.ResponseWriter, r *http.Request) {
	UrlName := r.URL.Query().Get("name")

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), strings.ToLower(UrlName))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error()))
		return
	}

	merchantInfo, err := m.Postgresdb.GetAllMerchantInfo(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while accessing merchant info: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, merchantInfo)
}

func (m *Merchant) MerchantSettingsInfoByOwner(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	settingsInfo, err := m.Postgresdb.GetMerchantSettingsInfo(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while accessing settings merchant info: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, settingsInfo)

}

func (m *Merchant) NewLocation(w http.ResponseWriter, r *http.Request) {
	type newLocation struct {
		Country           *string           `json:"country"`
		City              *string           `json:"city"`
		PostalCode        *string           `json:"postal_code"`
		Address           *string           `json:"address"`
		GeoPoint          location.GeoPoint `json:"geo_point"`
		PlaceId           *string           `json:"place_id"`
		FormattedLocation string            `json:"formatted_location"`
		IsPrimary         bool              `json:"is_primary"`
		IsActive          bool              `json:"is_active"`
	}
	var nl newLocation

	if err := validate.ParseStruct(r, &nl); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err := m.Postgresdb.NewLocation(r.Context(), database.Location{
		MerchantId:        employee.MerchantId,
		Country:           nl.Country,
		City:              nl.City,
		PostalCode:        nl.PostalCode,
		Address:           nl.Address,
		GeoPoint:          nl.GeoPoint,
		PlaceId:           nl.PlaceId,
		FormattedLocation: nl.FormattedLocation,
		IsPrimary:         nl.IsPrimary,
		IsActive:          nl.IsActive,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during adding location to database: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (m *Merchant) NewService(w http.ResponseWriter, r *http.Request) {
	type newConnectedProducts struct {
		ProductId  int `json:"id" validate:"required"`
		AmountUsed int `json:"amount_used" validate:"min=0,max=1000000"`
	}

	type newPhase struct {
		Name      string `json:"name" validate:"required"`
		Sequence  int    `json:"sequence" validate:"required"`
		Duration  int    `json:"duration" validate:"required,min=1,max=1440"`
		PhaseType string `json:"phase_type" validate:"required,eq=wait|eq=active"`
	}

	type newService struct {
		Name         string                   `json:"name" validate:"required"`
		Description  *string                  `json:"description"`
		Color        string                   `json:"color" validate:"required,hexcolor"`
		Price        *currencyx.Price         `json:"price"`
		Cost         *currencyx.Price         `json:"cost"`
		PriceNote    *string                  `json:"price_note"`
		CategoryId   *int                     `json:"category_id"`
		IsActive     bool                     `json:"is_active"`
		Settings     database.ServiceSettings `json:"settings"`
		Phases       []newPhase               `json:"phases" validate:"required"`
		UsedProducts []newConnectedProducts   `json:"used_products" validate:"required"`
	}
	var service newService

	if err := validate.ParseStruct(r, &service); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	if len(service.Phases) == 0 {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("service phases can not be empty"))
	}

	var dbPhases []database.ServicePhase
	durationSum := 0
	for _, phase := range service.Phases {
		dbPhases = append(dbPhases, database.ServicePhase{
			Id:        0,
			ServiceId: 0,
			Name:      phase.Name,
			Sequence:  phase.Sequence,
			Duration:  phase.Duration,
			PhaseType: phase.PhaseType,
		})
		durationSum += phase.Duration
	}

	var dbProducts []database.ConnectedProducts
	for _, product := range service.UsedProducts {
		dbProducts = append(dbProducts, database.ConnectedProducts{
			ProductId:  product.ProductId,
			ServiceId:  0,
			AmountUsed: product.AmountUsed,
		})
	}

	curr, err := m.Postgresdb.GetMerchantCurrency(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while getting merchant's currency: %s", err.Error()))
		return
	}

	if service.Price != nil {
		if service.Price.CurrencyCode() != curr {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("new service price's currency does not match merchant's currency"))
			return
		}
	}

	if service.Cost != nil {
		if service.Cost.CurrencyCode() != curr {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("new service cost's currency does not match merchant's currency"))
			return
		}
	}

	if err := m.Postgresdb.NewService(r.Context(), database.Service{
		Id:              0,
		MerchantId:      employee.MerchantId,
		CategoryId:      service.CategoryId,
		BookingType:     booking.Appointment,
		Name:            service.Name,
		Description:     service.Description,
		Color:           service.Color,
		TotalDuration:   durationSum,
		Price:           service.Price,
		Cost:            service.Cost,
		PriceNote:       service.PriceNote,
		IsActive:        service.IsActive,
		Sequence:        0,
		MinParticipants: 1,
		MaxParticipants: 1,
		ServiceSettings: service.Settings,
	}, dbPhases, dbProducts); err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error inserting service: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (m *Merchant) CheckUrl(w http.ResponseWriter, r *http.Request) {
	type merchantName struct {
		Name string `json:"merchant_name"`
	}
	var mn merchantName

	if err := validate.ParseStruct(r, &mn); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	urlName, err := validate.MerchantNameToUrlName(mn.Name)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error()))
		return
	}

	err = m.Postgresdb.IsMerchantUrlUnique(r.Context(), urlName)
	if err != nil {
		httputil.WriteJSON(w, http.StatusConflict, map[string]map[string]string{"error": {"message": err.Error(), "merchant_url": urlName}})
		return
	}

	merchantUrl := struct {
		Url string `json:"merchant_url"`
	}{
		Url: urlName,
	}

	httputil.Success(w, http.StatusOK, merchantUrl)
}

func (m *Merchant) GetAvailableTimes(w http.ResponseWriter, r *http.Request) {
	urlName := r.URL.Query().Get("name")
	urlStartDate := r.URL.Query().Get("start")
	urlEndDate := r.URL.Query().Get("end")

	startDate, err := time.Parse(time.RFC3339, urlStartDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid date format: %s", err.Error()))
		return
	}

	endDate, err := time.Parse(time.RFC3339, urlEndDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid date format: %s", err.Error()))
		return
	}

	startDate = startDate.UTC()
	endDate = endDate.UTC()

	urlServiceId, err := strconv.Atoi(r.URL.Query().Get("serviceId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("serviceId should be a number: %s", err.Error()))
		return
	}

	urlLocationId, err := strconv.Atoi(r.URL.Query().Get("locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("locationId should be a number: %s", err.Error()))
		return
	}

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), strings.ToLower(urlName))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error()))
		return
	}

	service, err := m.Postgresdb.GetServiceWithPhasesById(r.Context(), urlServiceId, merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving service: %s", err.Error()))
		return
	}

	if service.MerchantId != merchantId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("this service id does not belong to this merchant"))
		return
	}

	bookingSettings, err := m.Postgresdb.GetBookingSettingsByMerchantAndService(r.Context(), merchantId, service.Id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting booking settings for merchant: %s", err.Error()))
		return
	}

	merchantTz, err := m.Postgresdb.GetMerchantTimezoneById(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	reservedTimes, err := m.Postgresdb.GetReservedTimesForPeriod(r.Context(), merchantId, urlLocationId, startDate, endDate)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while calculating available time slots: %s", err.Error()))
		return
	}

	blockedTimes, err := m.Postgresdb.GetBlockedTimes(r.Context(), merchantId, startDate, endDate)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting blocked times for merchant: %s", err.Error()))
		return
	}

	businessHours, err := m.Postgresdb.GetBusinessHours(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting business hours: %s", err.Error()))
		return
	}

	now := time.Now()
	availableSlots := CalculateAvailableTimesPeriod(reservedTimes, blockedTimes, service.Phases, service.TotalDuration, bookingSettings.BufferTime, bookingSettings.BookingWindowMin, startDate, endDate, businessHours, now, merchantTz)

	httputil.Success(w, http.StatusOK, availableSlots)
}

func (m *Merchant) GetServices(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	services, err := m.Postgresdb.GetServicesByMerchantId(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving services for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, services)
}

func (m *Merchant) GetService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	serviceId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	service, err := m.Postgresdb.GetAllServicePageData(r.Context(), serviceId, employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving service for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, service)
}

func (m *Merchant) GetServiceFormOptions(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	formOptions, err := m.Postgresdb.GetServicePageFormOptions(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving service form options for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, formOptions)
}

func (m *Merchant) DeleteService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	serviceId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.DeleteServiceById(r.Context(), employee.MerchantId, serviceId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting service for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateService(w http.ResponseWriter, r *http.Request) {
	var pubServ database.ServiceWithPhasesAndSettings

	if err := validate.ParseStruct(r, &pubServ); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	serviceId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	if serviceId != pubServ.Id {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	if len(pubServ.Phases) == 0 {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("service phases can not be empty"))
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	durationSum := 0
	for _, phase := range pubServ.Phases {
		durationSum += phase.Duration
	}

	err = m.Postgresdb.UpdateServiceWithPhaseseById(r.Context(), database.ServiceWithPhasesAndSettings{
		Id:            pubServ.Id,
		MerchantId:    employee.MerchantId,
		CategoryId:    pubServ.CategoryId,
		Name:          pubServ.Name,
		Description:   pubServ.Description,
		Color:         pubServ.Color,
		TotalDuration: durationSum,
		Price:         pubServ.Price,
		Cost:          pubServ.Cost,
		PriceNote:     pubServ.PriceNote,
		IsActive:      pubServ.IsActive,
		Settings:      pubServ.Settings,
		Phases:        pubServ.Phases,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating service for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) DeactivateService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	serviceId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.DeactivateServiceById(r.Context(), employee.MerchantId, serviceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while deactivating service: %s", err.Error()))
		return
	}
}

func (m *Merchant) ActivateService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	serviceId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.ActivateServiceById(r.Context(), employee.MerchantId, serviceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while deactivating service: %s", err.Error()))
		return
	}
}

func (m *Merchant) ReorderServices(w http.ResponseWriter, r *http.Request) {
	type servicesOrder struct {
		CategoryId *int  `json:"category_id"`
		Services   []int `json:"services" validate:"required"`
	}

	var so servicesOrder

	err := validate.ParseStruct(r, &so)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.ReorderServices(r.Context(), employee.MerchantId, so.CategoryId, so.Services)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while ordering services: %s", err.Error()))
		return
	}
}

func (m *Merchant) GetCustomers(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	customers, err := m.Postgresdb.GetCustomersByMerchantId(r.Context(), employee.MerchantId, false)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving customers for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, customers)
}

func (m *Merchant) GetBlacklistedCustomers(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	customers, err := m.Postgresdb.GetCustomersByMerchantId(r.Context(), employee.MerchantId, true)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving blacklisted customers for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, customers)
}

func (m *Merchant) NewCustomer(w http.ResponseWriter, r *http.Request) {
	type newCustomer struct {
		FirstName   *string    `json:"first_name" validate:"required"`
		LastName    *string    `json:"last_name" validate:"required"`
		Email       *string    `json:"email"`
		PhoneNumber *string    `json:"phone_number"`
		Birthday    *time.Time `json:"birthday"`
		Note        *string    `json:"note"`
	}
	var customer newCustomer

	if err := validate.ParseStruct(r, &customer); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	customerId, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating customer id: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	if err := m.Postgresdb.NewCustomer(r.Context(), employee.MerchantId, database.Customer{
		Id:          customerId,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Email:       customer.Email,
		PhoneNumber: customer.PhoneNumber,
		Birthday:    customer.Birthday,
		Note:        customer.Note,
	}); err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error inserting customer for merchant: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (m *Merchant) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid customer id provided"))
		return
	}

	customerId, err := uuid.Parse(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting customer id to uuid: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.DeleteCustomerById(r.Context(), customerId, employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting customer for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	type Customer struct {
		Id          uuid.UUID  `json:"id" validate:"required,uuid"`
		FirstName   *string    `json:"first_name"`
		LastName    *string    `json:"last_name"`
		Email       *string    `json:"email"`
		PhoneNumber *string    `json:"phone_number"`
		Birthday    *time.Time `json:"birthday"`
		Note        *string    `json:"note"`
	}
	var customer Customer

	if err := validate.ParseStruct(r, &customer); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid customer id provided"))
		return
	}

	customerId, err := uuid.Parse(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting customer id to uuid: %s", err.Error()))
		return
	}

	if customerId != customer.Id {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid customer id provided"))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.UpdateCustomerById(r.Context(), employee.MerchantId, database.Customer{
		Id:          customer.Id,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Email:       customer.Email,
		PhoneNumber: customer.PhoneNumber,
		Birthday:    customer.Birthday,
		Note:        customer.Note,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating customer for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateMerchantFields(w http.ResponseWriter, r *http.Request) {
	var data database.MerchantSettingFields

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err := m.Postgresdb.UpdateMerchantFieldsById(r.Context(), employee.MerchantId, data)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating reservation fileds for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) GetPreferences(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	preferences, err := m.Postgresdb.GetPreferencesByMerchantId(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while accessing merchant preferences: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, preferences)
}

func (m *Merchant) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	var p database.PreferenceData

	if err := validate.ParseStruct(r, &p); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err := m.Postgresdb.UpdatePreferences(r.Context(), employee.MerchantId, p)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating preferences: %s", err.Error()))
		return
	}
}

func (m *Merchant) TransferCustomerBookings(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := uuid.Parse(fromStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error when parsing 'from' as uuid: %s", err.Error()))
		return
	}

	to, err := uuid.Parse(toStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error when parsing 'to' as uuid: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.TransferDummyBookings(r.Context(), employee.MerchantId, from, to)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while transfering bookings: %s", err.Error()))
		return
	}
}

func (m *Merchant) BlacklistCustomer(w http.ResponseWriter, r *http.Request) {
	type blacklistData struct {
		CustomerId      uuid.UUID `json:"id" validate:"required,uuid"`
		BlacklistReason *string   `json:"blacklist_reason"`
	}

	var data blacklistData

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid customer id provided"))
		return
	}

	customerId, err := uuid.Parse(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting customer id to uuid: %s", err.Error()))
		return
	}

	if customerId != data.CustomerId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid customer id provided"))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.SetBlacklistStatusForCustomer(r.Context(), employee.MerchantId, customerId, true, data.BlacklistReason)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while adding customer to blacklist: %s", err.Error()))
		return
	}
}

func (m *Merchant) UnBlacklistCustomer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid customer id provided"))
		return
	}

	customerId, err := uuid.Parse(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting customer id to uuid: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.SetBlacklistStatusForCustomer(r.Context(), employee.MerchantId, customerId, false, nil)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting customer from blacklist: %s", err.Error()))
		return
	}
}

func (m *Merchant) NewProduct(w http.ResponseWriter, r *http.Request) {
	type newProduct struct {
		Name          string           `json:"name" validate:"required"`
		Description   string           `json:"description"`
		Price         *currencyx.Price `json:"price"`
		Unit          string           `json:"unit" validate:"required"`
		MaxAmount     int              `json:"max_amount" validate:"min=0,max=10000000000"`
		CurrentAmount int              `json:"current_amount" validate:"min=0,max=10000000000"`
	}
	var product newProduct

	if err := validate.ParseStruct(r, &product); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	curr, err := m.Postgresdb.GetMerchantCurrency(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while getting merchant's currency: %s", err.Error()))
		return
	}

	if product.Price != nil {
		if product.Price.CurrencyCode() != curr {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("new product price's currency does not match merchant's currency"))
			return
		}
	}

	if err := m.Postgresdb.NewProduct(r.Context(), database.Product{
		Id:            0,
		MerchantId:    employee.MerchantId,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		Unit:          product.Unit,
		MaxAmount:     product.MaxAmount,
		CurrentAmount: product.CurrentAmount,
	}); err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error inserting product for merchant: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (m *Merchant) GetProducts(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	products, err := m.Postgresdb.GetProductsByMerchant(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving products for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, products)
}

func (m *Merchant) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid product id provided"))
		return
	}

	productId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting product id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.DeleteProductById(r.Context(), employee.MerchantId, productId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting product for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	type productData struct {
		Id            int              `json:"id"`
		Name          string           `json:"name" validate:"required"`
		Description   string           `json:"description"`
		Price         *currencyx.Price `json:"price"`
		Unit          string           `json:"unit" validate:"required"`
		MaxAmount     int              `json:"max_amount" validate:"min=0,max=10000000000"`
		CurrentAmount int              `json:"current_amount" validate:"min=0,max=10000000000"`
	}

	var prod productData

	if err := validate.ParseStruct(r, &prod); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid product id provided"))
		return
	}

	productId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting product id to int: %s", err.Error()))
		return
	}

	if productId != prod.Id {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid product id provided"))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.UpdateProduct(r.Context(), database.Product{
		Id:            prod.Id,
		MerchantId:    employee.MerchantId,
		Name:          prod.Name,
		Description:   prod.Description,
		Price:         prod.Price,
		Unit:          prod.Unit,
		MaxAmount:     prod.MaxAmount,
		CurrentAmount: prod.CurrentAmount,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating product for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) GetDisabledDaysForCalendar(w http.ResponseWriter, r *http.Request) {
	urlName := r.URL.Query().Get("name")

	urlServiceId, err := strconv.Atoi(r.URL.Query().Get("serviceId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("serviceId should be a number: %s", err.Error()))
		return
	}

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), strings.ToLower(urlName))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error()))
		return
	}

	bookingSettings, err := m.Postgresdb.GetBookingSettingsByMerchantAndService(r.Context(), merchantId, urlServiceId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving booking settings by merchant id: %s", err.Error()))
		return
	}

	merchantTz, err := m.Postgresdb.GetMerchantTimezoneById(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	now := time.Now().In(merchantTz)

	minDate := now.Add(time.Duration(bookingSettings.BookingWindowMin) * time.Minute)
	maxDate := now.AddDate(0, bookingSettings.BookingWindowMax, 0)

	businessHours, err := m.Postgresdb.GetNormalizedBusinessHours(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retrieving business hours by merchant id: %s", err.Error()))
		return
	}

	closedDays := []int{}

	for i := 0; i <= 6; i++ {
		if _, ok := businessHours[i]; !ok {
			closedDays = append(closedDays, i)
		}
	}

	type disabledDays struct {
		ClosedDays []int     `json:"closed_days"`
		MinDate    time.Time `json:"min_date"`
		MaxDate    time.Time `json:"max_date"`
	}

	httputil.Success(w, http.StatusOK, disabledDays{
		ClosedDays: closedDays,
		MinDate:    minDate,
		MaxDate:    maxDate,
	})
}

func (m *Merchant) GetBusinessHours(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	buseinessHours, err := m.Postgresdb.GetNormalizedBusinessHours(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving business hours by merchant id: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, buseinessHours)
}

func (m *Merchant) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	periodStr := r.URL.Query().Get("period")

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("timestamp could not be converted to date: %s", err.Error()))
		return
	}

	period, err := strconv.Atoi(periodStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("period could not be converted to int: %s", err.Error()))
		return
	}

	if period != 7 && period != 30 {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid period: %d", period))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	dashboardData, err := m.Postgresdb.GetDashboardData(r.Context(), employee.MerchantId, date, period)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving dashboard data: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, dashboardData)
}

func (m *Merchant) NewServiceCategory(w http.ResponseWriter, r *http.Request) {
	type newCategory struct {
		Name string `json:"name" validate:"required"`
	}
	var nc newCategory

	err := validate.ParseStruct(r, &nc)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.NewServiceCategory(r.Context(), employee.MerchantId, database.ServiceCategory{
		Name:     nc.Name,
		Sequence: 0,
	})
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while creating new service category %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateServiceCategory(w http.ResponseWriter, r *http.Request) {
	type categoryData struct {
		Name string `json:"name" validate:"required"`
	}

	var cd categoryData

	err := validate.ParseStruct(r, &cd)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")

	categoryId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.UpdateServiceCategoryById(r.Context(), employee.MerchantId, database.ServiceCategory{
		Id:   categoryId,
		Name: cd.Name,
	})
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while updating service category: %s", err.Error()))
		return
	}
}

func (m *Merchant) DeleteServiceCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	categoryId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting category id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.DeleteServiceCategoryById(r.Context(), employee.MerchantId, categoryId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while deleting service category: %s", err.Error()))
		return
	}
}

func (m *Merchant) ReorderServiceCategories(w http.ResponseWriter, r *http.Request) {
	type categoryOrder struct {
		Categories []int `json:"categories" validate:"required"`
	}

	var co categoryOrder

	err := validate.ParseStruct(r, &co)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.ReorderServiceCategories(r.Context(), employee.MerchantId, co.Categories)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while ordering services: %s", err.Error()))
		return
	}
}

// TODO: this does not check wether the service and product belong to the merchant updating it
func (m *Merchant) UpdateServiceProductConnections(w http.ResponseWriter, r *http.Request) {
	type updatedProductConnections struct {
		ProductId  int `json:"id" validate:"required"`
		AmountUsed int `json:"amount_used" validate:"min=0,max=1000000"`
	}

	type ProductData struct {
		ServiceId    int                         `json:"service_id" validate:"required"`
		UsedProducts []updatedProductConnections `json:"used_products" validate:"required"`
	}

	var updatedProducts ProductData

	if err := validate.ParseStruct(r, &updatedProducts); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	serviceId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	if serviceId != updatedProducts.ServiceId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	var dbProducts []database.ConnectedProducts
	for _, product := range updatedProducts.UsedProducts {
		dbProducts = append(dbProducts, database.ConnectedProducts{
			ProductId:  product.ProductId,
			ServiceId:  updatedProducts.ServiceId,
			AmountUsed: product.AmountUsed,
		})
	}

	err = m.Postgresdb.UpdateConnectedProducts(r.Context(), updatedProducts.ServiceId, dbProducts)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating products connected to service for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) GetCustomerStatistics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	customerId, err := uuid.Parse(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting customer id to uuid: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	customer, err := m.Postgresdb.GetCustomerStatsByMerchant(r.Context(), employee.MerchantId, customerId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving customer stats for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, customer)

}

func (m *Merchant) GetCustomerInfo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid customer id provided"))
		return
	}

	customerId, err := uuid.Parse(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting customer id to uuid: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	customer, err := m.Postgresdb.GetCustomerInfoByMerchant(r.Context(), employee.MerchantId, customerId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving customer info for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, customer)
}

func (m *Merchant) GetPublicServiceDetails(w http.ResponseWriter, r *http.Request) {
	urlName := strings.ToLower(chi.URLParam(r, "merchantName"))
	if urlName == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid merchant name provided"))
		return
	}

	serviceIdStr := chi.URLParam(r, "serviceId")
	if serviceIdStr == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	serviceId, err := strconv.Atoi(serviceIdStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	locationIdStr := chi.URLParam(r, "locationId")
	if locationIdStr == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid location id provided"))
		return
	}

	locationId, err := strconv.Atoi(locationIdStr)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting location id to int: %s", err.Error()))
		return
	}

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), urlName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error()))
		return
	}

	service, err := m.Postgresdb.GetServiceDetailsForMerchantPage(r.Context(), merchantId, serviceId, locationId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving service info: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, service)
}

func (m *Merchant) GetNextAvailable(w http.ResponseWriter, r *http.Request) {
	urlName := r.URL.Query().Get("name")

	urlServiceId, err := strconv.Atoi(r.URL.Query().Get("serviceId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("serviceId should be a number: %s", err.Error()))
		return
	}

	urlLocationId, err := strconv.Atoi(r.URL.Query().Get("locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("locationId should be a number: %s", err.Error()))
		return
	}

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), strings.ToLower(urlName))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error()))
		return
	}

	service, err := m.Postgresdb.GetServiceWithPhasesById(r.Context(), urlServiceId, merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving service: %s", err.Error()))
		return
	}

	bookingSettings, err := m.Postgresdb.GetBookingSettingsByMerchantAndService(r.Context(), merchantId, service.Id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting booking setting for merchant: %s", err.Error()))
		return
	}

	startDate := time.Now().In(time.UTC)
	endDate := startDate.AddDate(0, 3, 0)

	reservedTimes, err := m.Postgresdb.GetReservedTimesForPeriod(r.Context(), merchantId, urlLocationId, startDate, endDate)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while calculating available time slots: %s", err.Error()))
		return
	}

	blockedTimes, err := m.Postgresdb.GetBlockedTimes(r.Context(), merchantId, startDate, endDate)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting blocked times for merchant: %s", err.Error()))
		return
	}

	merchantTz, err := m.Postgresdb.GetMerchantTimezoneById(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	businessHours, err := m.Postgresdb.GetBusinessHours(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting business hours: %s", err.Error()))
		return
	}

	now := time.Now()
	availableSlots := CalculateAvailableTimesPeriod(reservedTimes, blockedTimes, service.Phases, service.TotalDuration, bookingSettings.BufferTime, bookingSettings.BookingWindowMin, startDate, endDate, businessHours, now, merchantTz)

	type nextAvailable struct {
		Date string `json:"date"`
		Time string `json:"time"`
	}

	var na nextAvailable

	for _, day := range availableSlots {
		if len(day.Morning) > 0 {
			na.Time = day.Morning[0]
			na.Date = day.Date
			break
		}
		if len(day.Afternoon) > 0 {
			na.Time = day.Afternoon[0]
			na.Date = day.Date
			break
		}
	}

	httputil.Success(w, http.StatusOK, na)
}

func (m *Merchant) DeleteMerchant(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err := m.Postgresdb.DeleteMerchant(r.Context(), employee.Id, employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) ChangeMerchantName(w http.ResponseWriter, r *http.Request) {
	type merchantName struct {
		Name string `json:"name" validate:"required"`
	}

	var data merchantName

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
	}

	urlName, err := validate.MerchantNameToUrlName(data.Name)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error()))
		return
	}

	err = m.Postgresdb.IsMerchantUrlUnique(r.Context(), urlName)
	if err != nil {
		httputil.Error(w, http.StatusConflict, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.ChangeMerchantNameAndURL(r.Context(), employee.MerchantId, data.Name, urlName)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating merchant's name: %s", err.Error()))
		return
	}
}

func (m *Merchant) NewBookingByMerchant(w http.ResponseWriter, r *http.Request) {
	type recurringRule struct {
		Frequency string `json:"frequency"`
		Interval  int    `json:"interval"`
		Weekday   string `json:"weekday"`
		Until     string `json:"until"`
		Mode      string `json:"mode"`
	}

	type customer struct {
		CustomerId  *uuid.UUID `json:"customer_id"`
		FirstName   *string    `json:"first_name"`
		LastName    *string    `json:"last_name"`
		Email       *string    `json:"email"`
		PhoneNumber *string    `json:"phone_number"`
	}

	type newBooking struct {
		BookingType  booking.Type   `json:"booking_type"`
		Customers    []customer     `json:"customers"`
		ServiceId    int            `json:"service_id"`
		LocationId   int            `json:"location_id"`
		TimeStamp    string         `json:"timestamp"`
		MerchantNote *string        `json:"merchant_note"`
		IsRecurring  bool           `json:"is_recurring"`
		Rrule        *recurringRule `json:"recurrence_rule"`
	}

	var nb newBooking

	if err := validate.ParseStruct(r, &nb); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	if nb.BookingType == booking.Appointment {
		if len(nb.Customers) > 1 {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("booking type does not match amount of customers"))
			return
		}
	} else {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("events/classes are not implemented yet"))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	merchantTz, err := m.Postgresdb.GetMerchantTimezoneById(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	bookedLocation, err := m.Postgresdb.GetLocationById(r.Context(), nb.LocationId, employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching location by this id: %s", err.Error()))
		return
	}

	service, err := m.Postgresdb.GetServiceWithPhasesById(r.Context(), nb.ServiceId, employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching service by this id: %s", err.Error()))
		return
	}

	// TODO: this should be a separate function
	// prevent null booking price and cost to avoid a lot of headaches
	var price currencyx.Price
	var cost currencyx.Price
	if service.Price == nil || service.Cost == nil {
		curr, err := m.Postgresdb.GetMerchantCurrency(r.Context(), employee.MerchantId)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while getting merchant's currency: %s", err.Error()))
			return
		}

		zeroAmount, err := currency.NewAmount("0", curr)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while creating new amount: %s", err.Error()))
		}

		if service.Price != nil {
			price = *service.Price
		} else {
			price = currencyx.Price{Amount: zeroAmount}
		}

		if service.Cost != nil {
			cost = *service.Cost
		} else {
			cost = currencyx.Price{Amount: zeroAmount}
		}
	} else {
		price = *service.Price
		cost = *service.Cost
	}

	timeStamp, err := time.Parse(time.RFC3339, nb.TimeStamp)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("timestamp could not be converted to time: %s", err.Error()))
		return
	}

	duration := time.Duration(service.TotalDuration) * time.Minute

	fromDate := timeStamp
	toDate := timeStamp.Add(duration)

	var customerId uuid.UUID

	if nb.Customers[0].CustomerId == nil {
		customerId, err = uuid.NewV7()
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating customer id: %s", err.Error()))
			return
		}

		if err := m.Postgresdb.NewCustomer(r.Context(), employee.MerchantId, database.Customer{
			Id:          customerId,
			FirstName:   nb.Customers[0].FirstName,
			LastName:    nb.Customers[0].LastName,
			Email:       nb.Customers[0].Email,
			PhoneNumber: nb.Customers[0].PhoneNumber,
			Birthday:    nil,
			Note:        nil,
		}); err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error inserting customer for merchant: %s", err.Error()))
			return
		}
	} else {
		customerId = *nb.Customers[0].CustomerId
	}

	var bookingId int

	if nb.IsRecurring {
		var freq rrule.Frequency

		switch strings.ToUpper(nb.Rrule.Frequency) {
		case "DAILY":
			freq = rrule.DAILY
		case "WEEKLY":
			freq = rrule.WEEKLY
		case "MONTHLY":
			freq = rrule.MONTHLY
		default:
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("recurring rule frequency is invalid"))
			return
		}

		untilTimeStamp, err := time.Parse(time.RFC3339, nb.Rrule.Until)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("until timestamp could not be converted to time: %s", err.Error()))
			return
		}

		var weekday rrule.Weekday
		switch strings.ToUpper(nb.Rrule.Weekday) {
		case rrule.MO.String():
			weekday = rrule.MO
		case rrule.TU.String():
			weekday = rrule.TU
		case rrule.WE.String():
			weekday = rrule.WE
		case rrule.TH.String():
			weekday = rrule.TH
		case rrule.FR.String():
			weekday = rrule.FR
		case rrule.SA.String():
			weekday = rrule.SA
		case rrule.SU.String():
			weekday = rrule.SU
		default:
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("incorrect weekday"))
			return
		}

		rrule, err := rrule.NewRRule(rrule.ROption{
			Freq:     freq,
			Dtstart:  fromDate,
			Interval: nb.Rrule.Interval,
			Wkst:     weekday,
			Until:    untilTimeStamp,
		})
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while creating rrule: %s", err.Error()))
			return
		}

		// recurring bookings have to be stored in local time and converted to UTC during generation
		fromDate = timeStamp.In(merchantTz)

		series, err := m.Postgresdb.NewBookingSeries(r.Context(), database.NewBookingSeries{
			BookingType:    booking.Appointment,
			MerchantId:     employee.MerchantId,
			EmployeeId:     employee.Id,
			ServiceId:      service.Id,
			LocationId:     bookedLocation.Id,
			Rrule:          rrule.String(),
			Dstart:         fromDate,
			Timezone:       merchantTz,
			PricePerPerson: price,
			CostPerPerson:  cost,
			Participants:   []uuid.UUID{customerId},
		})
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while creating new booking series: %s", err.Error()))
			return
		}

		bookingId, err = m.generateRecurringBookings(r.Context(), series, service.Phases)
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while generating recurring bookings: %s", err.Error()))
			return
		}
	} else {
		bookingId, err = m.Postgresdb.NewBookingByMerchant(r.Context(), database.NewMerchantBooking{
			Status:         booking.Booked,
			BookingType:    booking.Appointment,
			MerchantId:     employee.MerchantId,
			ServiceId:      service.Id,
			LocationId:     bookedLocation.Id,
			FromDate:       fromDate,
			ToDate:         toDate,
			MerchantNote:   nb.MerchantNote,
			PricePerPerson: price,
			CostPerPerson:  cost,
			CustomerId:     customerId,
			Phases:         service.Phases,
		})
		if err != nil {
			httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while creating new booking: %s", err.Error()))
			return
		}
	}

	customerEmail, err := m.Postgresdb.GetCustomerEmailById(r.Context(), employee.MerchantId, customerId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while getting customer's email: %s", err.Error()))
		return
	}

	urlName, err := m.Postgresdb.GetMerchantUrlNameById(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while getting merchant's url name: %s", err.Error()))
		return
	}

	toDateMerchantTz := toDate.In(merchantTz)
	fromDateMerchantTz := timeStamp.In(merchantTz)

	emailData := email.BookingConfirmationData{
		Time:        fromDateMerchantTz.Format("15:04") + " - " + toDateMerchantTz.Format("15:04"),
		Date:        fromDateMerchantTz.Format("Monday, January 2"),
		Location:    bookedLocation.FormattedLocation,
		ServiceName: service.Name,
		TimeZone:    merchantTz.String(),
		ModifyLink:  fmt.Sprintf("http://reservations.local:3000/m/%s/cancel/%d", urlName, bookingId),
	}

	lang := lang.LangFromContext(r.Context())

	err = email.BookingConfirmation(r.Context(), lang, customerEmail, emailData)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not send confirmation email for the booking: %s", err.Error()))
		return
	}

	hoursUntilBooking := time.Until(fromDateMerchantTz).Hours()

	if hoursUntilBooking >= 24 {

		reminderDate := fromDateMerchantTz.Add(-24 * time.Hour)
		email_id, err := email.BookingReminder(r.Context(), lang, customerEmail, emailData, reminderDate)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not schedule reminder email: %s", err.Error()))
			return
		}

		if email_id != "" { //check because return "" when email sending is off
			err = m.Postgresdb.UpdateEmailIdForBooking(r.Context(), bookingId, email_id)
			if err != nil {
				httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("failed to update email ID: %s", err.Error()))
				return
			}
		}
	}

	w.WriteHeader(http.StatusCreated)
}

func (m *Merchant) generateRecurringBookings(ctx context.Context, series database.CompleteBookingSeries, serivePhases []database.PublicServicePhase) (int, error) {
	tz, err := time.LoadLocation(series.Timezone)
	if err != nil {
		return 0, fmt.Errorf("error parsing location from booking series: %s", err.Error())
	}

	now := time.Now().UTC()
	end := now.AddDate(0, 3, 0)

	rrule, err := rrule.StrToRRule(series.Rrule)
	if err != nil {
		return 0, fmt.Errorf("error parsing rrule string: %s", err.Error())
	}

	occurrences := rrule.Between(now, end, true)

	existingOccurrences, err := m.Postgresdb.GetExistingOccurrenceDates(ctx, series.Id, now, end)
	if err != nil {
		return 0, fmt.Errorf("could not get existing occurrence dates: %s", err.Error())
	}

	existingMap := make(map[string]bool)
	for _, date := range existingOccurrences {
		existingMap[date.Format("2006-01-02")] = true
	}

	var duration time.Duration
	for _, phase := range serivePhases {
		duration += time.Duration(phase.Duration)
	}

	duration = duration * time.Minute

	var fromDates []time.Time
	var toDates []time.Time
	for _, date := range occurrences {
		if existingMap[date.Format("2006-01-02")] {
			continue
		}

		fromDate := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, tz)
		toDate := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, tz)
		toDate = toDate.Add(duration)

		fromDates = append(fromDates, fromDate.UTC())
		toDates = append(toDates, toDate.UTC())
	}

	return m.Postgresdb.BatchCreateRecurringBookings(ctx, database.NewRecurringBookings{
		BookingSeriesId: series.Id,
		BookingStatus:   booking.Booked,
		BookingType:     series.BookingType,
		MerchantId:      series.MerchantId,
		EmployeeId:      series.EmployeeId,
		ServiceId:       series.ServiceId,
		LocationId:      series.LocationId,
		FromDates:       fromDates,
		ToDates:         toDates,
		Phases:          serivePhases,
		Details:         series.Details,
		Participants:    series.Participants,
	})
}

func (m *Merchant) NewBlockedTime(w http.ResponseWriter, r *http.Request) {
	type newBlockedTime struct {
		Name string `json:"name" validate:"required"`
		// EmployeeIds []int  `json:"employee_ids" validate:"required"`
		FromDate string `json:"from_date" validate:"required"`
		ToDate   string `json:"to_date" validate:"required"`
		AllDay   bool   `json:"all_day"`
	}

	var nbt newBlockedTime

	if err := validate.ParseStruct(r, &nbt); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	fromDate, err := time.Parse(time.RFC3339, nbt.FromDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("fromDate timestamp could not be converted to time: %s", err.Error()))
		return
	}

	toDate, err := time.Parse(time.RFC3339, nbt.ToDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("toDate timestamp could not be converted to time: %s", err.Error()))
		return
	}

	if !toDate.After(fromDate) {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("toDate must be after fromDate"))
		return
	}

	err = m.Postgresdb.NewBlockedTime(r.Context(), employee.MerchantId, []int{employee.Id}, nbt.Name, fromDate, toDate, nbt.AllDay)
	// err = m.Postgresdb.NewBlockedTime(r.Context(), employee.MerchantId, nbt.EmployeeIds, nbt.Name, fromDate, toDate, nbt.AllDay)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not make new blocked time %s", err.Error()))
	}

	w.WriteHeader(http.StatusCreated)
}

func (m *Merchant) DeleteBlockedTime(w http.ResponseWriter, r *http.Request) {
	// type deleteData struct {
	// 	EmployeeId int `json:"employee_id" validate:"required"`
	// }

	// var dd deleteData

	// if err := validate.ParseStruct(r, &dd); err != nil {
	// 	httputil.Error(w, http.StatusBadRequest, err)
	// 	return
	// }
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	blockedTimeId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting service id to int: %s", err.Error()))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	err = m.Postgresdb.DeleteBlockedTime(r.Context(), blockedTimeId, employee.MerchantId, employee.Id)
	// err = m.Postgresdb.DeleteBlockedTime(r.Context(), blockedTimeId, employee.MerchantId, dd.EmployeeId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting blocked time for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateBlockedTime(w http.ResponseWriter, r *http.Request) {
	// employee id not ids but its a front end issue
	type blockedTimeData struct {
		Id   int    `json:"id" validate:"required"`
		Name string `json:"name" validate:"required"`
		// EmployeeId int    `json:"employee_id" validate:"required"`
		FromDate string `json:"from_date" validate:"required"`
		ToDate   string `json:"to_date" validate:"required"`
		AllDay   bool   `json:"all_day"`
	}

	var data blockedTimeData
	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid blocekd time id provided"))
		return
	}

	blockedTimeId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting blocked time id to int: %s", err.Error()))
		return
	}

	if blockedTimeId != data.Id {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid blocked time id provided"))
		return
	}

	employee := jwt.MustGetEmployeeFromContext(r.Context())

	fromDate, err := time.Parse(time.RFC3339, data.FromDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("fromDate timestamp could not be converted to time: %s", err.Error()))
		return
	}

	toDate, err := time.Parse(time.RFC3339, data.ToDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("toDate timestamp could not be converted to time: %s", err.Error()))
		return
	}

	if !toDate.After(fromDate) {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("toDate must be after fromDate"))
		return
	}

	err = m.Postgresdb.UpdateBlockedTime(r.Context(), database.BlockedTime{
		Id:         blockedTimeId,
		MerchantId: employee.MerchantId,
		// EmployeeId: data.EmployeeId,
		EmployeeId: employee.Id,
		Name:       data.Name,
		FromDate:   fromDate,
		ToDate:     toDate,
		AllDay:     data.AllDay,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating blocked time for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) GetEmployeesForCalendar(w http.ResponseWriter, r *http.Request) {
	employee := jwt.MustGetEmployeeFromContext(r.Context())

	employees, err := m.Postgresdb.GetEmployeesByMerchant(r.Context(), employee.MerchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving employees for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, employees)
}

func (m *Merchant) GetSummaryInfo(w http.ResponseWriter, r *http.Request) {
	urlName := r.URL.Query().Get("name")

	urlServiceId, err := strconv.Atoi(r.URL.Query().Get("serviceId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("serviceId should be a number: %s", err.Error()))
		return
	}

	urlLocationId, err := strconv.Atoi(r.URL.Query().Get("locationId"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("locationId should be a number: %s", err.Error()))
		return
	}

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), strings.ToLower(urlName))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving the merchant's id: %s", err.Error()))
		return
	}

	info, err := m.Postgresdb.GetMinimalServiceInfo(r.Context(), merchantId, urlServiceId, urlLocationId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving minimal service info: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, info)
}
