package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/booking"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Merchant struct {
	Postgresdb database.PostgreSQL
}

func (m *Merchant) InfoByName(w http.ResponseWriter, r *http.Request) {
	UrlName := r.URL.Query().Get("name")

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), strings.ToLower(UrlName))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving the merchant's id: %s", err.Error()))
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
	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	settingsInfo, err := m.Postgresdb.GetMerchantSettingsInfo(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while accessing settings merchant info: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, settingsInfo)

}

func (m *Merchant) NewLocation(w http.ResponseWriter, r *http.Request) {
	type newLocation struct {
		Country    string `json:"country" validate:"required"`
		City       string `json:"city" validate:"required"`
		PostalCode string `json:"postal_code" validate:"required"`
		Address    string `json:"address" validate:"required"`
	}
	var location newLocation

	if err := validate.ParseStruct(r, &location); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
		return
	}

	err = m.Postgresdb.NewLocation(r.Context(), database.Location{
		Id:         0,
		MerchantId: merchantId,
		Country:    location.Country,
		City:       location.City,
		PostalCode: location.PostalCode,
		Address:    location.Address,
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
		Name         string                 `json:"name" validate:"required"`
		Description  string                 `json:"description"`
		Color        string                 `json:"color" validate:"required,hexcolor"`
		Price        int                    `json:"price" validate:"min=0,max=1000000"`
		PriceNote    *string                `json:"price_note"`
		Cost         int                    `json:"cost" validate:"min=0,max=1000000"`
		CategoryId   *int                   `json:"category_id"`
		IsActive     bool                   `json:"is_active"`
		Phases       []newPhase             `json:"phases" validate:"required"`
		UsedProducts []newConnectedProducts `json:"used_products" validate:"required"`
	}
	var service newService

	if err := validate.ParseStruct(r, &service); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
		return
	}

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

	if err := m.Postgresdb.NewService(r.Context(), database.Service{
		Id:            0,
		MerchantId:    merchantId,
		CategoryId:    service.CategoryId,
		Name:          service.Name,
		Description:   service.Description,
		Color:         service.Color,
		TotalDuration: durationSum,
		Price:         service.Price,
		PriceNote:     service.PriceNote,
		Cost:          service.Cost,
		IsActive:      service.IsActive,
		Sequence:      0,
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

func (m *Merchant) GetHours(w http.ResponseWriter, r *http.Request) {
	urlName := r.URL.Query().Get("name")

	urlDay := r.URL.Query().Get("day")
	day, err := time.Parse("2006-01-02", urlDay)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid day format: %s", err.Error()))
		return
	}

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
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving the merchant's id: %s", err.Error()))
		return
	}

	service, err := m.Postgresdb.GetServiceWithPhasesById(r.Context(), urlServiceId, merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving service: %s", err.Error()))
		return
	}

	if service.MerchantId != merchantId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("this service id does not belong to this merchant"))
		return
	}

	reservedTimes, err := m.Postgresdb.GetReservedTimes(r.Context(), merchantId, urlLocationId, day)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while calculating available time slots: %s", err.Error()))
		return
	}

	dayOfWeek := int(day.Weekday())

	businessHours, err := m.Postgresdb.GetBusinessHoursByDay(r.Context(), merchantId, dayOfWeek)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting business hours for the day: %s", err.Error()))
		return
	}

	timezone, err := m.Postgresdb.GetMerchantTimezoneById(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while getting merchant's timezone: %s", err.Error()))
		return
	}

	merchantTz, err := time.LoadLocation(timezone)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while parsing merchant's timezone: %s", err.Error()))
		return
	}

	now := time.Now()
	availableSlots := booking.CalculateAvailableTimes(reservedTimes, service.Phases, service.TotalDuration, day, businessHours, now, merchantTz)

	httputil.Success(w, http.StatusOK, availableSlots)
}

func (m *Merchant) GetServices(w http.ResponseWriter, r *http.Request) {
	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	services, err := m.Postgresdb.GetServicesByMerchantId(r.Context(), merchantId, false)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving services for merchant: %s", err.Error()))
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	service, err := m.Postgresdb.GetAllServicePageData(r.Context(), serviceId, merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving service for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, service)
}

func (m *Merchant) GetServiceFormOptions(w http.ResponseWriter, r *http.Request) {
	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	formOptions, err := m.Postgresdb.GetServicePageFormOptions(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving service form options for merchant: %s", err.Error()))
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.DeleteServiceById(r.Context(), merchantId, serviceId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting service for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateService(w http.ResponseWriter, r *http.Request) {
	var pubServ database.PublicServiceWithPhases

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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	durationSum := 0
	for _, phase := range pubServ.Phases {
		durationSum += phase.Duration
	}

	err = m.Postgresdb.UpdateServiceWithPhaseseById(r.Context(), database.PublicServiceWithPhases{
		Id:            pubServ.Id,
		MerchantId:    merchantId,
		CategoryId:    pubServ.CategoryId,
		Name:          pubServ.Name,
		Description:   pubServ.Description,
		Color:         pubServ.Color,
		TotalDuration: durationSum,
		Price:         pubServ.Price,
		PriceNote:     pubServ.PriceNote,
		Cost:          pubServ.Cost,
		IsActive:      pubServ.IsActive,
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.DeactivateServiceById(r.Context(), merchantId, serviceId)
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.ActivateServiceById(r.Context(), merchantId, serviceId)
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.ReorderServices(r.Context(), merchantId, so.CategoryId, so.Services)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while ordering services: %s", err.Error()))
		return
	}
}

func (m *Merchant) GetCustomers(w http.ResponseWriter, r *http.Request) {
	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	customers, err := m.Postgresdb.GetCustomersByMerchantId(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving customers for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, customers)
}

func (m *Merchant) NewCustomer(w http.ResponseWriter, r *http.Request) {
	type newCustomer struct {
		FirstName   string `json:"first_name" validate:"required"`
		LastName    string `json:"last_name" validate:"required"`
		Email       string `json:"email" validate:"omitempty,email"`
		PhoneNumber string `json:"phone_number" validate:"omitempty,e164"`
	}
	var customer newCustomer

	if err := validate.ParseStruct(r, &customer); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	customerId, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating user id: %s", err.Error()))
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	if err := m.Postgresdb.NewCustomer(r.Context(), merchantId, database.Customer{
		Id:          customerId,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Email:       customer.Email,
		PhoneNumber: customer.PhoneNumber,
		IsDummy:     true,
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.DeleteCustomerById(r.Context(), customerId, merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting customer for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	type Customer struct {
		Id          uuid.UUID `json:"id" validate:"required,uuid"`
		FirstName   string    `json:"first_name" validate:"required"`
		LastName    string    `json:"last_name" validate:"required"`
		Email       string    `json:"email" validate:"omitempty,email"`
		PhoneNumber string    `json:"phone_number" validate:"omitempty,e164"`
	}
	var customer Customer

	if err := validate.ParseStruct(r, &customer); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

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

	if customerId != customer.Id {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid customer id provided"))
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.UpdateCustomerById(r.Context(), merchantId, database.Customer{
		Id:          customer.Id,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Email:       customer.Email,
		PhoneNumber: customer.PhoneNumber,
		IsDummy:     true,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating customer for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateMerchantFields(w http.ResponseWriter, r *http.Request) {
	type MerchantFileds struct {
		Introduction  string                      `json:"introduction"`
		Announcement  string                      `json:"announcement"`
		AboutUs       string                      `json:"about_us"`
		ParkingInfo   string                      `json:"parking_info"`
		PaymentInfo   string                      `json:"payment_info"`
		BusinessHours map[int][]database.TimeSlot `json:"business_hours"`
	}
	var data MerchantFileds

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.UpdateMerchantFieldsById(r.Context(), merchantId, data.Introduction, data.Announcement, data.AboutUs, data.PaymentInfo, data.ParkingInfo, data.BusinessHours)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating reservation fileds for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	preferences, err := m.Postgresdb.GetPreferencesByMerchantId(r.Context(), merchantId)
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.UpdatePreferences(r.Context(), merchantId, p)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating preferences: %s", err.Error()))
		return
	}

}

func (m *Merchant) TransferCustomerApps(w http.ResponseWriter, r *http.Request) {
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.TransferDummyAppointments(r.Context(), merchantId, from, to)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while transfering appointments: %s", err.Error()))
		return
	}
}

func (m *Merchant) BlacklistCustomer(w http.ResponseWriter, r *http.Request) {
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.AddCustomerToBlacklist(r.Context(), merchantId, customerId)
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.RemoveCustomerFromBlacklist(r.Context(), merchantId, customerId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting customer from blacklist: %s", err.Error()))
		return
	}
}

func (m *Merchant) NewProduct(w http.ResponseWriter, r *http.Request) {
	type newProduct struct {
		Name          string `json:"name" validate:"required"`
		Description   string `json:"description"`
		Price         int    `json:"price" validate:"required,min=0,max=1000000"`
		Unit          string `json:"unit" validate:"required"`
		MaxAmount     int    `json:"max_amount" validate:"min=0,max=10000000000"`
		CurrentAmount int    `json:"current_amount" validate:"min=0,max=10000000000"`
	}
	var product newProduct

	if err := validate.ParseStruct(r, &product); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
		return
	}

	if err := m.Postgresdb.NewProduct(r.Context(), database.Product{
		Id:            0,
		MerchantId:    merchantId,
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
	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	products, err := m.Postgresdb.GetProductsByMerchant(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving products for merchant: %s", err.Error()))
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.DeleteProductById(r.Context(), merchantId, productId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while deleting product for merchant: %s", err.Error()))
		return
	}
}

func (m *Merchant) UpdateProduct(w http.ResponseWriter, r *http.Request) {

	type productData struct {
		Id            int    `json:"id"`
		Name          string `json:"name" validate:"required"`
		Description   string `json:"description"`
		Price         int    `json:"price" validate:"required,min=0,max=1000000"`
		Unit          string `json:"unit" validate:"required"`
		MaxAmount     int    `json:"max_amount" validate:"min=0,max=10000000000"`
		CurrentAmount int    `json:"current_amount" validate:"min=0,max=10000000000"`
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.UpdateProduct(r.Context(), database.Product{
		Id:            prod.Id,
		MerchantId:    merchantId,
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

func (m *Merchant) GetClosedDays(w http.ResponseWriter, r *http.Request) {
	urlName := r.URL.Query().Get("name")

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), strings.ToLower(urlName))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving the merchant's id: %s", err.Error()))
		return
	}

	businessHours, err := m.Postgresdb.GetNormalizedBusinessHours(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving business hours by merchant id: %s", err.Error()))
		return
	}

	closedDays := []int{}

	for i := 0; i <= 6; i++ {
		if _, ok := businessHours[i]; !ok {
			closedDays = append(closedDays, i)
		}
	}

	httputil.Success(w, http.StatusOK, closedDays)
}

func (m *Merchant) GetBusinessHours(w http.ResponseWriter, r *http.Request) {
	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	buseinessHours, err := m.Postgresdb.GetNormalizedBusinessHours(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving business hours by merchant id: %s", err.Error()))
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	dashboardData, err := m.Postgresdb.GetDashboardData(r.Context(), merchantId, date, period)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving dashboard data: %s", err.Error()))
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.NewServiceCategory(r.Context(), merchantId, database.ServiceCategory{
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
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting product id to int: %s", err.Error()))
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.UpdateServiceCategoryById(r.Context(), merchantId, database.ServiceCategory{
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
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting product id to int: %s", err.Error()))
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.DeleteServiceCategoryById(r.Context(), merchantId, categoryId)
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

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.ReorderServiceCategories(r.Context(), merchantId, co.Categories)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while ordering services: %s", err.Error()))
		return
	}
}

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
