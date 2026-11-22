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

	// The *All permissions grant access to every employee's records,
	// without them the actor is limited to their own
	TeamViewAll     = Permission{"team:view_all"}
	CalendarViewAll = Permission{"calendar:view_all"}
	BookingsAll     = Permission{"bookings:all"}
	BlockedTimesAll = Permission{"blocked_times:all"}
	AvatarsAll      = Permission{"avatars:all"}
)

var rolePermissions = map[types.EmployeeRole][]Permission{
	types.EmployeeRoleOwner: {
		MerchantDelete, MerchantRename, SettingsEdit, LocationImagesEdit, TeamManage, TeamViewAll, CalendarViewAll,
		BookingsAll, BlockedTimesAll, AvatarsAll,
	},
	types.EmployeeRoleAdmin: {
		SettingsEdit, LocationImagesEdit, TeamManage, TeamViewAll, CalendarViewAll,
		BookingsAll, BlockedTimesAll, AvatarsAll,
	},
	types.EmployeeRoleStaff: {},
}

func ForRole(role types.EmployeeRole) []Permission {
	return slices.Clone(rolePermissions[role])
}

func Has(a actor.EmployeeContext, p Permission) bool {
	return slices.Contains(rolePermissions[a.Role], p)
}

// Scope of employee data the actor can access
// nil when the actor has the permission for every employee
func EmployeeScope(a actor.EmployeeContext, all Permission) *int {
	if Has(a, all) {
		return nil
	}

	employeeId := a.EmployeeId
	return &employeeId
}

func CanActFor(a actor.EmployeeContext, all Permission, employeeIds ...int) bool {
	if Has(a, all) {
		return true
	}

	for _, employeeId := range employeeIds {
		if employeeId != a.EmployeeId {
			return false
		}
	}

	return true
}
