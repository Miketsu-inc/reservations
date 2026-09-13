package merchantpage

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/frontend/apps/tango"
)

type Handler struct {
	service  *merchantServ.Service
	renderer *pageRenderer
}

func NewHandler(service *merchantServ.Service) *Handler {
	tangoDist, _ := tango.StaticFilesPath()

	return &Handler{
		service:  service,
		renderer: mustNewRenderer(tangoDist, config.LoadEnvVars().TANGO_URL),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	merchantName := chi.URLParam(r, "merchantName")
	if merchantName == "" {
		h.writeNotFound(w)
		return
	}

	merchantInfo, err := h.service.GetInfo(r.Context(), merchantName)
	if err != nil {
		if errors.Is(err, merchantServ.ErrMerchantNotFound) {
			h.writeNotFound(w)
			return
		}

		slog.ErrorContext(r.Context(), "get merchant page data", "merchant", merchantName, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	pageHTML, err := h.renderer.render(merchantInfo)
	if err != nil {
		slog.ErrorContext(r.Context(), "render merchant page", "merchant", merchantName, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	httputil.WriteHTML(w, http.StatusOK, pageHTML)
}

func (h *Handler) writeNotFound(w http.ResponseWriter) {
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	httputil.WriteHTML(w, http.StatusNotFound, h.renderer.indexHTML)
}
