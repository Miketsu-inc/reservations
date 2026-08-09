package middleware

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/logger"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

func (m *Manager) RequestID(next http.Handler) httputil.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}

		l := logger.FromContext(r.Context()).With(
			slog.String("request_id", id),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		ctx := logger.WithContext(r.Context(), l)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))

		return nil
	}
}
