package integrations

import (
	"fmt"
	"net/http"

	"github.com/miketsu-inc/reservations/backend/cmd/config"
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

func (h *Handler) Routes() *httputil.Router {
	r := httputil.NewRouter()

	r.Get("/google/calendar/callback", h.GoogleCalendarCallback)
	r.Post("/google/calendar/watch", h.GoogleCalendarWatch)

	return r
}

const calendarStateCookie = "calendar-oauth-state"

var calendarCookieScope = oauthutil.CookieScope{Path: "/api/v1/integrations/google/calendar/callback"}

func (h *Handler) GoogleCalendarCallback(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	stateStr := r.URL.Query().Get("state")

	if err := oauthutil.ValidateState(r, calendarStateCookie); err != nil {
		return err
	}
	oauthutil.DeleteCookies(w, calendarCookieScope, calendarStateCookie)

	err := h.service.GoogleCalendarCallback(r.Context(), code, stateStr)
	if err != nil {
		return externalcalendarServ.ErrStatus.Resolve(err, "GoogleCalendarCallback")
	}

	// TEMP for testing environment
	http.Redirect(w, r, fmt.Sprintf("%s/integrations", config.LoadEnvVars().JABULANI_URL), http.StatusPermanentRedirect)

	return nil
}

// This is called by google for notification about a calendar change.
// It should return 200 even on internal failure as returning 200
// indicates to google that the server recived the notification.
// Any errors should be logged
func (h *Handler) GoogleCalendarWatch(w http.ResponseWriter, r *http.Request) error {
	channelId := r.Header.Get("X-Goog-Channel-ID")
	resourceId := r.Header.Get("X-Goog-Resource-ID")
	state := r.Header.Get("X-Goog-Resource-State")

	// Initial handshake notification
	if state == "sync" {
		return nil
	}

	h.service.GoogleCalendarWatch(r.Context(), channelId, resourceId)

	return nil
}
