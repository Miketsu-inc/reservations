package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
)

type Appointment struct {
	Postgresdb database.PostgreSQL
}

func (a *Appointment) Create(w http.ResponseWriter, r *http.Request) {
	var app database.Appointment

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&app); err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := a.Postgresdb.NewAppointment(app); err != nil {
		slog.Debug(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
