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
// allowes them to access the http route, should be called after the jwt middleware
func SubscriptionMiddleware(tiers ...subscription.Tier) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			employee := jwt.MustGetEmployeeFromContext(r.Context())

			db := database.New()

			tier, err := db.GetMerchantSubscriptionTier(r.Context(), employee.MerchantId)
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
