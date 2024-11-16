package handlers

import (
	"fmt"
	"net/http"

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
		FromDate     string `json:"from_date" validate:"required"`
		ToDate       string `json:"to_date" validate:"required"`
	}
	var newApp NewAppointment

	if err := validate.ParseStruct(r, &newApp); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := a.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
		return
	}

	app := database.Appointment{
		Id:         0,
		ClientId:   userID,
		MerchantId: merchantId,
		ServiceId:  newApp.ServiceId,
		LocationId: newApp.LocationId,
		FromDate:   newApp.FromDate,
		ToDate:     newApp.ToDate,
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
