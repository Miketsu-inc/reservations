package sub

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/middlewares/jwt"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
	"github.com/miketsu-inc/reservations/backend/pkg/subscription"
)

// Subscription middleware that check's if the merchant subscription tier
// allowes them to access the http route, should not be called before the jwt middleware
func SubscriptionMiddleware(tiers ...subscription.Tier) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userId := jwt.UserIDFromContext(r.Context())

			db := database.New()

			merchantId, err := db.GetMerchantIdByOwnerId(r.Context(), userId)
			if err != nil {
				httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error while retrieving merchant from owner id: %s", err.Error()))
				return
			}

			tier, err := db.GetMerchantSubscriptionTier(r.Context(), merchantId)
			if err != nil {
				httputil.Error(w, http.StatusBadRequest, fmt.Errorf("error during getting merchant's subscription tier: %s", err.Error()))
				return
			}

			if !slices.Contains(tiers, tier) {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("this resource can only be accessed with %s tiers", tiers))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
