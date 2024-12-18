package handlers

import (
	"fmt"
	"net/http"
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
	}
	var newApp NewAppointment

	if err := validate.ParseStruct(r, &newApp); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByUrlName(r.Context(), newApp.MerchantName)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no merchant found by this name: %s", err.Error()))
		return
	}

	service, err := a.Postgresdb.GetServiceById(r.Context(), newApp.ServiceId)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no service found by this id: %s", err.Error()))
		return
	}

	duration, err := time.ParseDuration(service.Duration + "m")
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

	app := database.Appointment{
		Id:         0,
		ClientId:   userID,
		MerchantId: merchantId,
		ServiceId:  newApp.ServiceId,
		LocationId: newApp.LocationId,
		FromDate:   from_date,
		ToDate:     to_date,
	}
	if err := a.Postgresdb.NewAppointment(r.Context(), app); err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("could not make new apppointment: %v", err))
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
