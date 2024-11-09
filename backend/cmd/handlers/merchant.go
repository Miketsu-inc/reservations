package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/validate"
)

type Merchant struct {
	Postgresdb database.PostgreSQL
}

func (m *Merchant) MerchantByName(w http.ResponseWriter, r *http.Request) {
	UrlName := r.URL.Query().Get("name")

	merchant, err := m.Postgresdb.GetMerchantByUrlName(r.Context(), strings.ToLower(UrlName))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error while retriving merchant: %s", err.Error()))
		return
	}

	utils.WriteJSON(w, http.StatusOK, merchant)
}

func (m *Merchant) NewLocation(w http.ResponseWriter, r *http.Request) {
	type newLocation struct {
		Country    string `json:"country" validate:"required"`
		City       string `json:"city" validate:"required"`
		PostalCode string `json:"postal_code" validate:"required"`
		Address    string `json:"address" validate:"required"`
	}
	var location newLocation

	if err := utils.ParseJSON(r, &location); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unexpected error during json parsing: %s", err.Error()))
		return
	}

	if errors := validate.Struct(location); errors != nil {
		utils.WriteJSON(w, http.StatusBadRequest, map[string]map[string]string{"errors": errors})
		return
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
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
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error during adding location to database: %s", err.Error()))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"success": "Location added to merchant successfully"})
}

func (m *Merchant) NewService(w http.ResponseWriter, r *http.Request) {
	type newService struct {
		Name     string `json:"name" validate:"required"`
		Duration string `json:"duration" validate:"required"`
		Price    string `json:"price" validate:"required"`
	}
	var services []newService

	//might fail when there's only 1 service needs testing
	if err := utils.ParseJSON(r, &services); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("unexpected error during json parsing: %s", err.Error()))
		return
	}

	for _, service := range services {
		if errors := validate.Struct(service); errors != nil {
			utils.WriteJSON(w, http.StatusBadRequest, map[string]map[string]string{"errors": errors})
			return
		}
	}

	userID := jwt.UserIDFromContext(r.Context())

	merchantId, err := m.Postgresdb.GetMerchantIdByOwnerId(r.Context(), userID)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("no merchant found for this user: %s", err.Error()))
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
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error inserting services: %s", err.Error()))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"success": "Service added to merchant successfully"})
}
