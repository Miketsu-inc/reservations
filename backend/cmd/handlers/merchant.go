package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
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
