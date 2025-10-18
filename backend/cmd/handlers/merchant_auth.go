package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/lang"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/subscription"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type MerchantAuth struct {
	Postgresdb database.PostgreSQL
}

func (m *MerchantAuth) Signup(w http.ResponseWriter, r *http.Request) {
	type signUpData struct {
		Name         string `json:"name" validate:"required"`
		ContactEmail string `json:"contact_email" validate:"required,email"`
		Timezone     string `json:"timezone" validate:"required,timezone"`
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

	userID := jwt.MustGetUserIDFromContext(r.Context())

	merchantID, err := uuid.NewV7()
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating merchant id: %s", err.Error()))
		return
	}

	language := lang.LangFromContext(r.Context())
	curr := currencyx.FindBest(language)

	err = m.Postgresdb.NewMerchant(r.Context(), userID, database.Merchant{
		Id:               merchantID,
		Name:             signup.Name,
		UrlName:          urlName,
		ContactEmail:     signup.ContactEmail,
		Introduction:     "",
		Announcement:     "",
		AboutUs:          "",
		ParkingInfo:      "",
		PaymentInfo:      "",
		Timezone:         signup.Timezone,
		CurrencyCode:     curr,
		SubscriptionTier: subscription.Free,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error creating a merchant: %s", err.Error()))
		return
	}

	businessHours := map[int][]database.TimeSlot{
		0: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		1: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		2: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		3: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		4: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		5: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
		6: {{StartTime: "09:00:00", EndTime: "17:00:00"}},
	}

	err = m.Postgresdb.UpdateBusinessHours(r.Context(), merchantID, businessHours)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during creating business hours for merchant: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
