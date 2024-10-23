package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
)

type Appointment struct {
	Postgresdb database.PostgreSQL
}

func (a *Appointment) Create(w http.ResponseWriter, r *http.Request) {
	type NewAppointment struct {
		MerchantName string `json:"merchant_name" validate:"required"`
		TypeName     string `json:"type_name" validate:"required"`
		LocationName string `json:"location_name" validate:"required"`
		FromDate     string `json:"from_date" validate:"required"`
		ToDate       string `json:"to_date" validate:"required"`
	}
	var newApp NewAppointment

	if err := utils.ParseJSON(r, &newApp); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if errors := utils.StructValidation(newApp); errors != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]map[string]string{"error": errors})
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDCtxKey).(uuid.UUID)
	if !ok {
		panic("Authenticated route called without jwt user id")
	}

	app := database.Appointment{
		Id:           0,
		ClientId:     userID,
		MerchantName: newApp.MerchantName,
		TypeName:     newApp.TypeName,
		LocationName: newApp.LocationName,
		FromDate:     newApp.FromDate,
		ToDate:       newApp.ToDate,
	}
	if err := a.Postgresdb.NewAppointment(r.Context(), app); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, "Could not make new apppointment")
		return
	}

}

func (a *Appointment) GetEvents(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	apps, err := a.Postgresdb.GetAppointmentsByMerchant(r.Context(), "Hair salon", start, end)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if len(apps) == 0 {
		println("No appointments found")
	}

	utils.WriteJSON(w, http.StatusOK, apps)
}
