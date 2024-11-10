package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Merchant struct {
	Postgresdb database.PostgreSQL
}

func (m *Merchant) MerchantByName(w http.ResponseWriter, r *http.Request) {
	UrlName := r.URL.Query().Get("name")

	merchantId, err := m.Postgresdb.GetMerchantIdByUrlName(r.Context(), strings.ToLower(UrlName))
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving the merchant's id: %s", err.Error()))
		return
	}

	merchant, err := m.Postgresdb.GetMerchantById(r.Context(), merchantId)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("error while retriving merchant by id: %s", err.Error()))
		return
	}

	httputil.Success(w, http.StatusOK, merchant)
}

func (m *Merchant) NewLocation(w http.ResponseWriter, r *http.Request) {
	type newLocation struct {
		Country    string `json:"country" validate:"required"`
		City       string `json:"city" validate:"required"`
		PostalCode string `json:"postal_code" validate:"required"`
		Address    string `json:"address" validate:"required"`
	}
	var location newLocation

	if err := validate.ParseStruct(r, &location); err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
		return
	}

	err = m.Postgresdb.NewLocation(r.Context(), database.Location{
		Id:         0,
		MerchantId: merchantId,
		Country:    location.Country,
		City:       location.City,
		PostalCode: location.PostalCode,
		Address:    location.Address,
	})
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during adding location to database: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (m *Merchant) NewService(w http.ResponseWriter, r *http.Request) {
	type newService struct {
		Name     string `json:"name" validate:"required"`
		Duration string `json:"duration" validate:"required"`
		Price    string `json:"price" validate:"required"`
	}
	var services []newService

	// TODO: how should ParseStruct handle this?
	// -----
	if err := httputil.ParseJSON(r, &services); err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("unexpected error during json parsing: %s", err.Error()))
		return
	}

	for _, service := range services {
		if err := validate.Struct(service); err != nil {
			httputil.Error(w, http.StatusBadRequest, err)
			return
		}
	}
	// -----

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
		return
	}

	dbServices := make([]database.Service, len(services))
	for i, svcs := range services {
		dbServices[i] = database.Service{
			Id:         0,
			MerchantId: merchantId,
			Name:       svcs.Name,
			Duration:   svcs.Duration,
			Price:      svcs.Price,
		}
	}

	if err := m.Postgresdb.NewServices(r.Context(), dbServices); err != nil {
		httputil.Error(w, http.StatusInternalServerError, fmt.Errorf("unexpected error inserting services: %s", err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
}
