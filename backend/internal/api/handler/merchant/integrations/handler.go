package integrations

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	externalcalendarServ "github.com/miketsu-inc/reservations/backend/internal/service/externalcalendar"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/oauthutil"
)

type Handler struct {
	service *externalcalendarServ.Service
}

func NewHandler(s *externalcalendarServ.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/google/calendar", h.GoogleCalendar)
	r.Put("/google/calendar/callback", h.GoogleCalendarCallback)
	r.Post("/google/calendar/watch", h.GoogleCalendarWatch)

	return r
}

func (h *Handler) GoogleCalendar(w http.ResponseWriter, r *http.Request) {
	url, state, err := h.service.GoogleCalendar(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	oauthutil.SetOauthStateCookie(w, state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handler) GoogleCalendarCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if err := oauthutil.ValidateOauthState(r); err != nil {
		httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during oauth state validation: %s", err.Error()))
		return
	}

	err := h.service.GoogleCalendarCallback(r.Context(), code)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, err)
		return
	}

	// TEMP for testing environment
	http.Redirect(w, r, "http://app.reservations.local:3000/integrations", http.StatusPermanentRedirect)
}

// This is called by google for notification about a calendar change.
// It should return 200 even on internal failure as returning 200
// indicates to google that the server recived the notification.
// Any errors should be logged
func (h *Handler) GoogleCalendarWatch(w http.ResponseWriter, r *http.Request) {
	channelId := r.Header.Get("X-Goog-Channel-ID")
	resourceId := r.Header.Get("X-Goog-Resource-ID")
	state := r.Header.Get("X-Goog-Resource-State")

	// Initial handshake notification
	if state == "sync" {
		return
	}

	h.service.GoogleCalendarWatch(r.Context(), channelId, resourceId)
}
