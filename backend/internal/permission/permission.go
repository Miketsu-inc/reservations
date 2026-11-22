package permission

import (
	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

// Staff only work with their own calendar, owners and admins with the whole team's
func CanManageTeam(a actor.EmployeeContext) bool {
	return a.Role != types.EmployeeRoleStaff
}

// The employee whose data reads are limited to, nil when the actor sees the whole team
func EmployeeScope(a actor.EmployeeContext) *int {
	if CanManageTeam(a) {
		return nil
	}

	employeeId := a.EmployeeId
	return &employeeId
}

func CanManageEmployees(a actor.EmployeeContext, employeeIds ...int) bool {
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
