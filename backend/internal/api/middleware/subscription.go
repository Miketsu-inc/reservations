package middleware

import (
	"net/http"
	"slices"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

var ErrSubscriptionAccessDenied = &apperr.Error{Code: "subscription_access_denied", Message: "you do not have access to this resource"}

// Subscription middleware that check's if the merchant subscription tier
// allowes them to access the http route, should be called after the authentication middleware
func (m *Manager) Subscription(tiers ...types.SubTier) func(next http.Handler) httputil.HandlerFunc {
	return func(next http.Handler) httputil.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {

			actor := actor.MustGetFromContext(r.Context())

			tier, err := m.merchantRepo.GetMerchantSubscriptionTier(r.Context(), actor.MerchantId)
			if err != nil {
				return err
			}

			if !slices.Contains(tiers, tier) {
				return &apperr.APIError{
					Status: http.StatusUnauthorized,
					Err:    ErrSubscriptionAccessDenied,
					Meta:   map[string]any{"required_tiers": tiers},
				}
			}

			next.ServeHTTP(w, r)

			return nil
		}
	}
}
