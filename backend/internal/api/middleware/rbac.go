package middleware

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/jwt"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/httputil"
)

// Role based access control middleware that check's wether an employee can access
// a resource based on their role, should be called after the authentication middleware
func (m *Manager) RoleBasedAccessControl(roles ...types.EmployeeRole) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			employee, ok := jwt.GetEmployeeFromContext(r.Context())
			if !ok {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("you need an employee account to access this"))
				return
			}

			if !slices.Contains(roles, employee.Role) {
				httputil.Error(w, http.StatusUnauthorized, fmt.Errorf("this resource can only be accessed with %s roles", roles))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
