package handlers

import (
	"log/slog"
	"net/http"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
)

type Appointment struct {
	Postgresdb database.PostgreSQL
}

func (a *Appointment) Create(w http.ResponseWriter, r *http.Request) {
	var app database.Appointment

	if err := utils.ParseJSON(r, &app); err != nil {
		slog.Error(err.Error())
		http.Error(w, "JSON parsing failed on appointment", http.StatusInternalServerError)
		return
	}

	if err := a.Postgresdb.NewAppointment(r.Context(), app); err != nil {
		slog.Error(err.Error())
		http.Error(w, "Could not make new apppointment", http.StatusInternalServerError)
		return
	}

	}

func (a *Appointment) GetEvents(w http.ResponseWriter, r *http.Request) {
	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	// id := r.URL.Query().Get("id")

	apps, err := a.Postgresdb.GetAppointmentsByMerchant(r.Context(), "Hair salon", start, end)
	if err != nil {
		println("the error is here")
		slog.Error(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if len(apps) == 0 {
		println("No appointments found")
	}

	utils.WriteJSON(w, http.StatusOK, apps)
}
