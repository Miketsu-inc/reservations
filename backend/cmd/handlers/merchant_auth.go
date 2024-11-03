package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type MerchantAuth struct {
	Postgresdb database.PostgreSQL
}

func (m *MerchantAuth) Signup(w http.ResponseWriter, r *http.Request) {
	type signUpData struct {
		Name         string `json:"name" validate:"required"`
		ContactEmail string `json:"contact_email" validate:"required,email"`
	}
	var signup signUpData

	if errors := validate.Struct(signup); errors != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]map[string]string{"errors": errors})
		return
	}

	if err := utils.ParseJSON(r, &signup); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unexpected error during json parsing: %s", err.Error()))
		return
	}

	urlName, err := validate.MerchantNameToUrlName(signup.Name)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error()))
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDCtxKey).(uuid.UUID)
	assert.True(ok, "Authenticated route called without jwt user id", r.Context().Value(middlewares.UserIDCtxKey), userID)

	merchantID, err := uuid.NewV7()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating merchant id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.NewMerchant(r.Context(), database.Merchant{
		Id:           merchantID,
		Name:         signup.Name,
		UrlName:      urlName,
		OwnerId:      userID,
		ContactEmail: signup.ContactEmail,
		Settings:     make(map[string]bool),
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during adding merchant to database: %s", err.Error()))
		return
	}
}
