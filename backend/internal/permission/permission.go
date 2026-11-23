package permission

import (
	"slices"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

var roleGrants = map[types.EmployeeRole]Grants{
	types.EmployeeRoleOwner: {
		Actions: []Action{DeleteMerchant, RenameMerchant, EditSettings, ManageTeam, ManageCustomers, ManageServices},
		View:    map[Resource]Scope{Bookings: All, BlockedTimes: All, Team: All, Services: All, Reviews: All},
		Edit:    map[Resource]Scope{Bookings: All, BlockedTimes: All, Team: Own, Services: All, Reviews: All},
		Manage:  []types.EmployeeRole{types.EmployeeRoleAdmin, types.EmployeeRoleStaff},
	},
	types.EmployeeRoleAdmin: {
		Actions: []Action{EditSettings, ManageTeam, ManageCustomers, ManageServices},
		View:    map[Resource]Scope{Bookings: All, BlockedTimes: All, Team: All, Services: All, Reviews: All},
		Edit:    map[Resource]Scope{Bookings: All, BlockedTimes: All, Team: Own, Services: All, Reviews: All},
		Manage:  []types.EmployeeRole{types.EmployeeRoleStaff},
	},
	types.EmployeeRoleStaff: {
		View: map[Resource]Scope{Bookings: Own, BlockedTimes: Own, Team: Own, Services: All, Reviews: Own},
		Edit: map[Resource]Scope{Bookings: Own, BlockedTimes: Own, Team: Own, Services: Own, Reviews: Own},
	},
}

func ForRole(role types.EmployeeRole) Grants {
	return roleGrants[role]
}

type Checker struct {
	employeeId int
	grants     Grants
}

func For(a actor.EmployeeContext) Checker {
	return Checker{employeeId: a.EmployeeId, grants: roleGrants[a.Role]}
}

func (c Checker) Do(action Action) bool {
	return slices.Contains(c.grants.Actions, action)
}

func (c Checker) ViewAll(resource Resource) bool {
	return c.grants.View[resource] == All
}

func (c Checker) View(resource Resource, employeeIds ...int) bool {
	if len(employeeIds) == 0 {
		return false
	}

	switch c.grants.View[resource] {
	case All:
		return true
	case Own:
		return slices.Contains(employeeIds, c.employeeId)
	default:
		return false
	}
}

// Filter for repository queries, nil when every employee's records are visible
func (c Checker) ViewFilter(resource Resource) *int {
	switch c.grants.View[resource] {
	case All:
		return nil
	case Own:
		employeeId := c.employeeId
		return &employeeId
	default:
		assert.Never("view filter for a resource the role can't view", resource)

		noEmployee := 0
		return &noEmployee
	}
}

func (c Checker) EditAll(resource Resource) bool {
	return c.grants.Edit[resource] == All
}

func (c Checker) Edit(resource Resource, employeeIds ...int) bool {
	if len(employeeIds) == 0 {
		return false
	}

	switch c.grants.Edit[resource] {
	case All:
		return true
	case Own:
		for _, employeeId := range employeeIds {
			if employeeId != c.employeeId {
				return false
			}
		}

		return true
	default:
		return false
	}
}

func (c Checker) Manage(role types.EmployeeRole) bool {
	return slices.Contains(c.grants.Manage, role)
}
