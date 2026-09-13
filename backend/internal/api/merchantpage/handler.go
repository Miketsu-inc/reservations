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
		renderer: newRenderer(tangoDist, config.LoadEnvVars().TANGO_URL),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	merchantName := chi.URLParam(r, "merchantName")
	if merchantName == "" {
		err := h.writeNotFound(w)
		if err != nil {
			slog.ErrorContext(r.Context(), "render merchant page metadata", "merchant", merchantName, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		return
	}

	info, err := h.service.GetInfo(r.Context(), merchantName)
	if err != nil {
		if errors.Is(err, merchantServ.ErrMerchantNotFound) {
			err := h.writeNotFound(w)
			if err != nil {
				slog.ErrorContext(r.Context(), "render merchant page metadata", "merchant", merchantName, "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			return
		}

		slog.ErrorContext(r.Context(), "render merchant page", "merchant", merchantName, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	document, err := h.renderer.render(info)
	if err != nil {
		slog.ErrorContext(r.Context(), "render merchant page metadata", "merchant", merchantName, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = httputil.WriteHTML(w, http.StatusOK, document)
	if err != nil {
		slog.ErrorContext(r.Context(), "render merchant page metadata", "merchant", merchantName, "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) writeNotFound(w http.ResponseWriter) error {
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	return httputil.WriteHTML(w, http.StatusNotFound, h.renderer.indexHTML)
}
