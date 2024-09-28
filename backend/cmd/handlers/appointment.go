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

	w.WriteHeader(http.StatusOK)
}
