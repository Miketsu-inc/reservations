package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
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

	if err := validate.ParseStruct(r, &signup); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
	}

	urlName, err := validate.MerchantNameToUrlName(signup.Name)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("unexpected error during merchant url name conversion: %s", err.Error()))
		return
	}

	err = m.Postgresdb.IsMerchantUrlUnique(r.Context(), urlName)
	if err != nil {
		httputil.Error(w, http.StatusConflict, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantID, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating merchant id: %s", err.Error()))
		return
	}

	err = m.Postgresdb.NewMerchant(r.Context(), database.Merchant{
		Id:           merchantID,
		Name:         signup.Name,
		UrlName:      urlName,
		OwnerId:      userID,
		ContactEmail: signup.ContactEmail,
		Settings:     database.MerchantSettings{},
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during adding merchant to database: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
