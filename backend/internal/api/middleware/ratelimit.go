package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

var ErrRateLimited = &apperr.Error{Code: "rate_limited", Message: "too many requests, please try again later"}

// Every use of the middleware has it's own limit, should be called after the authentication middleware
func (m *Manager) RateLimitByUser(requestLimit int, window time.Duration) func(next http.Handler) httputil.HandlerFunc {
	limiter := httprate.NewRateLimiter(requestLimit, window)

	return func(next http.Handler) httputil.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			userId := jwt.MustGetUserIDFromContext(r.Context())

			if limiter.OnLimit(w, r, userId.String()) {
				return &apperr.APIError{
					Status: http.StatusTooManyRequests,
					Err:    ErrRateLimited,
				}
			}

			next.ServeHTTP(w, r)

			return nil
		}
	}
}
