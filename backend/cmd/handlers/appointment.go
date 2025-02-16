package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
		UserComment  string `json:"user_comment"`
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
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("timestamp could not be converted to int: %s", err.Error()))
		return
	}

	const postgresTimeFormat = "2006-01-02T15:04:05-0700"

	toDate := timeStamp.Add(duration)
	from_date := timeStamp.Format(postgresTimeFormat)
	to_date := toDate.Format(postgresTimeFormat)

	if err := a.Postgresdb.NewAppointment(r.Context(), database.Appointment{
		Id:              0,
		UserId:          userID,
		MerchantId:      merchantId,
		ServiceId:       newApp.ServiceId,
		LocationId:      newApp.LocationId,
		FromDate:        from_date,
		ToDate:          to_date,
		UserComment:     newApp.UserComment,
		MerchantComment: "",
		PriceThen:       service.Price,
		CostThen:        service.Cost,
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

func (a *Appointment) UpdateMerchantComment(w http.ResponseWriter, r *http.Request) {
	// validate:"required" on MerchantComment would fail
	// if an empty string arrives as a comment
	type newComment struct {
		Id              int    `json:"id" validate:"required"`
		MerchantComment string `json:"merchant_comment"`
	}

	var data newComment

	if err := validate.ParseStruct(r, &data); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while searching for the merchant by this owner id: %s", err.Error()))
		return
	}

	if err := a.Postgresdb.UpdateMerchantCommentById(r.Context(), merchantId, data.Id, data.MerchantComment); err != nil {
		httputil.Error(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
