package products

import (
	"net/http"

	productServ "github.com/miketsu-inc/reservations/backend/internal/service/product"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

type lowStockProductResp struct {
	Id            int     `json:"id"`
	Name          string  `json:"name"`
	MaxAmount     int     `json:"max_amount"`
	CurrentAmount int     `json:"current_amount"`
	Unit          string  `json:"unit"`
	FillRatio     float64 `json:"fill_ratio"`
}

func (h *Handler) GetLowStock(w http.ResponseWriter, r *http.Request) error {
	products, err := h.service.GetLowStock(r.Context())
	if err != nil {
		return productServ.ErrStatus.Resolve(err, "GetLowStock")
	}

	httputil.Success(w, http.StatusOK, mapToGetLowStockResp(products))

	return nil
}
