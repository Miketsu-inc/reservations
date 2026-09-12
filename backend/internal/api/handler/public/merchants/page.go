package merchants

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	merchantServ "github.com/miketsu-inc/reservations/backend/internal/service/merchant"
)

type merchantInfoService interface {
	GetInfo(ctx context.Context, merchantName string) (domain.MerchantInfo, error)
}

type PageHandler struct {
	service  merchantInfoService
	renderer *pageRenderer
}

func NewPageHandler(service merchantInfoService, dist fs.FS, publicBaseURL string) (*PageHandler, error) {
	renderer, err := newPageRenderer(dist, publicBaseURL)
	if err != nil {
		return nil, err
	}

	return &PageHandler{
		service:  service,
		renderer: renderer,
	}, nil
}

func (h *PageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	merchantName := chi.URLParam(r, "merchantName")
	if merchantName == "" {
		h.writeNotFound(w)
		return
	}

	info, err := h.service.GetInfo(r.Context(), merchantName)
	if err != nil {
		if errors.Is(err, merchantServ.ErrMerchantNotFound) {
			h.writeNotFound(w)
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

	writeHTML(w, http.StatusOK, document)
}

func (h *PageHandler) writeNotFound(w http.ResponseWriter) {
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	writeHTML(w, http.StatusNotFound, h.renderer.indexDocument)
}

func writeHTML(w http.ResponseWriter, status int, document []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(document)
}
