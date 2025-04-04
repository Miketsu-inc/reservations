package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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
	type newService struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		Color       string `json:"color" validate:"required,hexcolor"`
		Duration    int    `json:"duration" validate:"required,min=1,max=1440"`
		Price       int    `json:"price" validate:"min=0,max=1000000"`
		Cost        int    `json:"cost" validate:"min=0,max=1000000"`
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

	if err := m.Postgresdb.NewService(r.Context(), database.Service{
		Id:          0,
		MerchantId:  merchantId,
		Name:        service.Name,
		Description: service.Description,
		Color:       service.Color,
		Duration:    service.Duration,
		Price:       service.Price,
		Cost:        service.Cost,
	}); err != nil {
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

	service, err := m.Postgresdb.GetServiceById(r.Context(), urlServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving service: %s", err.Error()))
		return
	}

	if service.MerchantId != merchantId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("this serivce id does not belong to this merchant"))
		return
	}

	reservedTimes, err := m.Postgresdb.GetReservedTimes(r.Context(), merchantId, urlLocationId, day)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while calculating available time slots: %s", err.Error()))
		return
	}

	availableSlots := calculateAvailableTimes(reservedTimes, service.Duration, day)

	httputil.Success(w, http.StatusOK, availableSlots)
}

type FormattedAvailableTimes struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
}

func calculateAvailableTimes(reserved []database.AppointmentTime, serviceDuration int, day time.Time) FormattedAvailableTimes {
	year, month, day_ := day.Date()
	location := day.Location()

	businessStart := time.Date(year, month, day_, 8, 0, 0, 0, location)
	businessEnd := time.Date(year, month, day_, 17, 0, 0, 0, location)

	duration := time.Duration(serviceDuration) * time.Minute
	current := businessStart

	morning := []string{}
	afternoon := []string{}
	for current.Add(duration).Before(businessEnd) || current.Add(duration).Equal(businessEnd) {
		timeEnd := current.Add(duration)
		available := true

		for _, appt := range reserved {
			if timeEnd.After(appt.From_date) && current.Before(appt.To_date) {
				current = appt.To_date
				timeEnd = current.Add(duration)
				available = false
				break
			}
		}

		if available && timeEnd.Before(businessEnd) || timeEnd.Equal(businessEnd) {
			formattedTime := fmt.Sprintf("%02d:%02d", current.Hour(), current.Minute())

			if current.Hour() < 12 {
				morning = append(morning, formattedTime)
			} else if current.Hour() >= 12 {
				afternoon = append(afternoon, formattedTime)
			}

			current = timeEnd
		}
	}

	availableTimes := FormattedAvailableTimes{
		Morning:   morning,
		Afternoon: afternoon,
	}

	return availableTimes
}

func (m *Merchant) GetServices(w http.ResponseWriter, r *http.Request) {
	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	services, err := m.Postgresdb.GetServicesByMerchantId(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving services for merchant: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, services)
}

func (m *Merchant) DeleteService(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	serviceId, err := strconv.Atoi(id)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting serivce id to int: %s", err.Error()))
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
	var serv database.PublicService

	if err := validate.ParseStruct(r, &serv); err != nil {
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
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting serivce id to int: %s", err.Error()))
		return
	}

	if serviceId != serv.Id {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid service id provided"))
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.UpdateServiceById(r.Context(), database.Service{
		Id:          serv.Id,
		MerchantId:  merchantId,
		Name:        serv.Name,
		Description: serv.Description,
		Color:       serv.Color,
		Duration:    serv.Duration,
		Price:       serv.Price,
		Cost:        serv.Cost,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while updating service for merchant: %s", err.Error()))
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
		Introduction  string                       `json:"introduction"`
		Announcement  string                       `json:"announcement"`
		AboutUs       string                       `json:"about_us"`
		ParkingInfo   string                       `json:"parking_info"`
		PaymentInfo   string                       `json:"payment_info"`
		BusinessHours map[int][]database.TimeSlots `json:"business_hours"`
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
		StockQuantity int    `json:"stock_quantity" validate:"required,min=0,max=10000"`
		UsagePerUnit  int    `json:"usage_per_unit" validate:"min=0,max=10000"`
		ServiceIds    []int  `json:"service_ids"`
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
		StockQuantity: product.StockQuantity,
		UsagePerUnit:  product.UsagePerUnit,
		ServiceIds:    product.ServiceIds,
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

	services, err := m.Postgresdb.GetServicesByMerchantId(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving services for merchant: %s", err.Error()))
		return
	}

	filteredServices := make([]map[string]interface{}, len(services))
	for i, service := range services {
		filteredServices[i] = map[string]interface{}{
			"Id":    service.Id,
			"Name":  service.Name,
			"Color": service.Color,
		}
	}

	response := map[string]interface{}{
		"products": products,
		"services": filteredServices,
	}

	httputil.Success(w, http.StatusOK, response)
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

	var prod database.PublicProduct

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
		StockQuantity: prod.StockQuantity,
		UsagePerUnit:  prod.UsagePerUnit,
		ServiceIds:    prod.ServiceIds,
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
