package permission

import (
	"encoding/json"
	"slices"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type Permission struct {
	name string
}

func (p Permission) String() string {
	return p.name
}

func (p Permission) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.name)
}

var (
	MerchantDelete     = Permission{"merchant:delete"}
	MerchantRename     = Permission{"merchant:rename"}
	SettingsEdit       = Permission{"settings:edit"}
	LocationImagesEdit = Permission{"location_images:edit"}
	TeamManage         = Permission{"team:manage"}

	// Only checked through their EmployeeScope
	teamViewAll     = Permission{"team:view_all"}
	calendarViewAll = Permission{"calendar:view_all"}
	bookingsAll     = Permission{"bookings:all"}
	blockedTimesAll = Permission{"blocked_times:all"}
	avatarsAll      = Permission{"avatars:all"}
)

var rolePermissions = map[types.EmployeeRole][]Permission{
	types.EmployeeRoleOwner: {
		MerchantDelete, MerchantRename, SettingsEdit, LocationImagesEdit, TeamManage,
		teamViewAll, calendarViewAll, bookingsAll, blockedTimesAll, avatarsAll,
	},
	types.EmployeeRoleAdmin: {
		SettingsEdit, LocationImagesEdit, TeamManage,
		teamViewAll, calendarViewAll, bookingsAll, blockedTimesAll, avatarsAll,
	},
	types.EmployeeRoleStaff: {},
}

func ForRole(role types.EmployeeRole) []Permission {
	return rolePermissions[role]
}

func Has(a actor.EmployeeContext, p Permission) bool {
	return slices.Contains(rolePermissions[a.Role], p)
}

// Limits an actor to their own records unless they have the scope's permission
type EmployeeScope struct {
	all Permission
}

var (
	Team         = EmployeeScope{teamViewAll}
	Calendar     = EmployeeScope{calendarViewAll}
	Bookings     = EmployeeScope{bookingsAll}
	BlockedTimes = EmployeeScope{blockedTimesAll}
	Avatars      = EmployeeScope{avatarsAll}
)

func (s EmployeeScope) AllowsEveryone(a actor.EmployeeContext) bool {
	return Has(a, s.all)
}

func (s EmployeeScope) Allows(a actor.EmployeeContext, employeeIds ...int) bool {
	if s.AllowsEveryone(a) {
		return true
	}

	for _, employeeId := range employeeIds {
		if employeeId != a.EmployeeId {
			return false
		}
	}

	return true
}

// Employee filter for repository queries, nil when the actor is allowed everyone
func (s EmployeeScope) EmployeeFilter(a actor.EmployeeContext) *int {
	if s.AllowsEveryone(a) {
		return nil
	}

	employeeId := a.EmployeeId
	return &employeeId
}
