package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
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
		Name     string `json:"name" validate:"required"`
		Duration string `json:"duration" validate:"required"`
		Price    string `json:"price" validate:"required"`
	}
	var services []newService

	// TODO: how should ParseStruct handle this?
	// -----
	if err := httputil.ParseJSON(r, &services); err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("unexpected error during json parsing: %s", err.Error()))
		return
	}

	for _, service := range services {
		if err := validate.Struct(service); err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}
	}
	// -----

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
		return
	}

	dbServices := make([]database.Service, len(services))
	for i, svcs := range services {
		dbServices[i] = database.Service{
			Id:         0,
			MerchantId: merchantId,
			Name:       svcs.Name,
			Duration:   svcs.Duration,
			Price:      svcs.Price,
		}
	}

	if err := m.Postgresdb.NewServices(r.Context(), dbServices); err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error inserting services: %s", err.Error()))
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

	duration, err := strconv.Atoi(service.Duration)
	assert.Nil(err, "Service duration from database could not be converted to int", service.Duration, duration)

	availableSlots := calculateAvailableTimes(reservedTimes, duration, day)

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
