package middleware

import (
	"net/http"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/permission"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

var ErrEmployeeAccountRequired = &apperr.Error{Code: "employee_account_required", Message: "you need an employee account to access this"}
var ErrPermissionDenied = &apperr.Error{Code: "permission_denied", Message: "you do not have access to this resource"}

// Checks that the employee has every given permission,
// should be called after the employee authentication middleware
func (m *Manager) RequirePermission(perms ...permission.Permission) func(next http.Handler) httputil.HandlerFunc {
	return func(next http.Handler) httputil.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {

			actor, ok := actor.GetFromContext(r.Context())
			if !ok {
				return &apperr.APIError{
					Status: http.StatusUnauthorized,
					Err:    ErrEmployeeAccountRequired,
				}
			}

			for _, perm := range perms {
				if !permission.Has(actor, perm) {
					return &apperr.APIError{
						Status: http.StatusForbidden,
						Err:    ErrPermissionDenied,
						Meta:   map[string]any{"required_permissions": perms},
					}
				}
			}

			next.ServeHTTP(w, r)

			return nil
		}
	}
}
