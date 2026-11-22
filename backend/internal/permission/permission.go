package permission

import (
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

func CanManageTeam(a actor.EmployeeContext) bool {
	return a.Role != types.EmployeeRoleStaff
}

// Scope of data the actor can access
// nil when the actor sees the whole team
func EmployeeScope(a actor.EmployeeContext) *int {
	if CanManageTeam(a) {
		return nil
	}

	employeeId := a.EmployeeId
	return &employeeId
}

func CanActFor(a actor.EmployeeContext, employeeIds ...int) bool {
	if CanManageTeam(a) {
		return true
	}

	for _, employeeId := range employeeIds {
		if employeeId != a.EmployeeId {
			return false
		}
	}

	return true
}
