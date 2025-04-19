package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Appointment struct {
	Postgresdb database.PostgreSQL
}

func (a *Appointment) Create(w http.ResponseWriter, r *http.Request) {
	type NewAppointment struct {
		MerchantName string `json:"merchant_name" validate:"required"`
		ServiceId    int    `json:"service_id" validate:"required"`
		LocationId   int    `json:"location_id" validate:"required"`
		TimeStamp    string `json:"timeStamp" validate:"required"`
		UserTz       string `json:"user_tz" validate:"required,timezone"`
		UserNote     string `json:"user_note"`
	}
	var newApp NewAppointment

	if err := validate.ParseStruct(r, &newApp); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByUrlName(r.Context(), newApp.MerchantName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching merchant by this name: %s", err.Error()))
		return
	}

	err = a.Postgresdb.IsUserBlacklisted(r.Context(), merchantId, userID)
	if err != nil {
		httputil.Error(w, http.StatusForbidden, err)
		return
	}

	service, err := a.Postgresdb.GetServiceById(r.Context(), newApp.ServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching service by this id: %s", err.Error()))
		return
	}

	if service.MerchantId != merchantId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("this serivce id does not belong to this merchant"))
		return
	}

	location, err := a.Postgresdb.GetLocationById(r.Context(), newApp.LocationId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching location by this id: %s", err.Error()))
		return
	}

	if location.MerchantId != merchantId {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("this location id does not belong to this merchant"))
		return
	}

	duration, err := time.ParseDuration(strconv.Itoa(service.Duration) + "m")
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("duration could not be parsed: %s", err.Error()))
		return
	}

	timeStamp, err := time.Parse(time.RFC3339, newApp.TimeStamp)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("timestamp could not be converted to time: %s", err.Error()))
		return
	}

	const postgresTimeFormat = "2006-01-02T15:04:05-0700"

	toDate := timeStamp.Add(duration)
	from_date := timeStamp.Format(postgresTimeFormat)
	to_date := toDate.Format(postgresTimeFormat)

	if err := a.Postgresdb.NewAppointment(r.Context(), database.Appointment{
		Id:           0,
		UserId:       userID,
		MerchantId:   merchantId,
		ServiceId:    newApp.ServiceId,
		LocationId:   newApp.LocationId,
		FromDate:     from_date,
		ToDate:       to_date,
		UserNote:     newApp.UserNote,
		MerchantNote: "",
		PriceThen:    service.Price,
		CostThen:     service.Cost,
	}); err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not make new apppointment: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *Appointment) GetAppointments(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	apps, err := a.Postgresdb.GetAppointmentsByMerchant(r.Context(), merchantId, start, end)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	httputil.Success(w, http.StatusOK, apps)
}

func (a *Appointment) CancelAppointmentByMerchant(w http.ResponseWriter, r *http.Request) {
	type cancelAppointmentData struct {
		CancellationReason string `json:"cancellation_reason"`
	}

	urlId := chi.URLParam(r, "id")

	if urlId == "" {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("invalid appointment id provided"))
		return
	}

	appId, err := strconv.Atoi(urlId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while converting appointment id to int: %s", err.Error()))
		return
	}

	var cancelData cancelAppointmentData

	if err := validate.ParseStruct(r, &cancelData); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userId := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retriving merchant from owner id: %s", err.Error()))
		return
	}

	err = a.Postgresdb.CancelAppointmentByMerchant(r.Context(), merchantId, appId, cancelData.CancellationReason)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while cancelling appointment by merchant: %s", err.Error()))
		return
	}
}

func (a *Appointment) UpdateAppointmentData(w http.ResponseWriter, r *http.Request) {
	// validate:"required" on MerchantNote would fail
	// if an empty string arrives as a note
	type appointmentData struct {
		Id           int    `json:"id" validate:"required"`
		MerchantNote string `json:"merchant_note"`
		FromDate     string `json:"from_date" validate:"required"`
		ToDate       string `json:"to_date" validate:"required"`
	}

	var appData appointmentData

	if err := validate.ParseStruct(r, &appData); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching for the merchant by this owner id: %s", err.Error()))
		return
	}

	fromDate, err := time.Parse(time.RFC3339, appData.FromDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("from_date could not be converted to time: %s", err.Error()))
		return
	}

	toDate, err := time.Parse(time.RFC3339, appData.ToDate)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("to_date could not be converted to time: %s", err.Error()))
		return
	}

	const postgresTimeFormat = "2006-01-02T15:04:05-0700"

	formattedFromDate := fromDate.Format(postgresTimeFormat)
	formattedToDate := toDate.Format(postgresTimeFormat)

	if err := a.Postgresdb.UpdateAppointmentData(r.Context(), merchantId, appData.Id, appData.MerchantNote, formattedFromDate, formattedToDate); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}
}
