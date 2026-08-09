package middleware

import (
	"net/http"
	"slices"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/apperr"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

var ErrEmployeeAccountRequired = &apperr.Error{Code: "employee_account_required", Message: "you need an employee account to access this"}
var ErrRoleAccessDenied = &apperr.Error{Code: "role_access_denied", Message: "you do not have access to this resource"}

// Role based access control middleware that check's wether an employee can access
// a resource based on their role, should be called after the authentication middleware
func (m *Manager) RoleBasedAccessControl(roles ...types.EmployeeRole) func(next http.Handler) httputil.HandlerFunc {
	return func(next http.Handler) httputil.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {

			actor, ok := actor.GetFromContext(r.Context())
			if !ok {
				return &apperr.APIError{
					Status: http.StatusUnauthorized,
					Err:    ErrEmployeeAccountRequired,
				}
			}

			if !slices.Contains(roles, actor.Role) {
				return &apperr.APIError{
					Status: http.StatusUnauthorized,
					Err:    ErrRoleAccessDenied,
					Meta:   map[string]any{"required_roles": roles},
				}
			}

			next.ServeHTTP(w, r)

			return nil
		}
	}
}
