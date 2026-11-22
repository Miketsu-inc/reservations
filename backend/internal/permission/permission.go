package permission

import (
	"encoding/json"
	"slices"

	"github.com/miketsu-inc/reservations/backend/internal/api/middleware/actor"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

// Merchant wide operations, an employee either can or can't do them
type Action struct {
	name string
}

func (a Action) String() string {
	return a.name
}

func (a Action) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.name)
}

var (
	MerchantDelete     = Action{"merchant:delete"}
	MerchantRename     = Action{"merchant:rename"}
	SettingsEdit       = Action{"settings:edit"}
	LocationImagesEdit = Action{"location_images:edit"}
	TeamManage         = Action{"team:manage"}
)

// Records that belong to employees, access depends on whose they are
type Resource struct {
	name string
}

func (r Resource) String() string {
	return r.name
}

// Text marshaling so resources can be JSON object keys
func (r Resource) MarshalText() ([]byte, error) {
	return []byte(r.name), nil
}

var (
	Bookings     = Resource{"bookings"}
	BlockedTimes = Resource{"blocked_times"}
	Team         = Resource{"team"}
	Calendar     = Resource{"calendar"}
	Avatars      = Resource{"avatars"}
)

// The zero value means no access
type Scope struct {
	name string
}

func (s Scope) String() string {
	return s.name
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.name)
}

var (
	Own = Scope{"own"}
	All = Scope{"all"}
)

type Grants struct {
	Actions []Action
	View    map[Resource]Scope
	Edit    map[Resource]Scope
	Manage  []types.EmployeeRole
}

var roleGrants = map[types.EmployeeRole]Grants{
	types.EmployeeRoleOwner: {
		Actions: []Action{MerchantDelete, MerchantRename, SettingsEdit, LocationImagesEdit, TeamManage},
		View:    map[Resource]Scope{Bookings: All, BlockedTimes: All, Team: All, Calendar: All},
		Edit:    map[Resource]Scope{Bookings: All, BlockedTimes: All, Avatars: Own},
		Manage:  []types.EmployeeRole{types.EmployeeRoleAdmin, types.EmployeeRoleStaff},
	},
	types.EmployeeRoleAdmin: {
		Actions: []Action{SettingsEdit, LocationImagesEdit, TeamManage},
		View:    map[Resource]Scope{Bookings: All, BlockedTimes: All, Team: All, Calendar: All},
		Edit:    map[Resource]Scope{Bookings: All, BlockedTimes: All, Avatars: Own},
		Manage:  []types.EmployeeRole{types.EmployeeRoleStaff},
	},
	types.EmployeeRoleStaff: {
		View: map[Resource]Scope{Bookings: Own, BlockedTimes: Own, Team: Own, Calendar: Own},
		Edit: map[Resource]Scope{Bookings: Own, BlockedTimes: Own, Avatars: Own},
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

// A record is visible when any of its employees is visible, so shared records show up for each of them
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

		// Employee ids are serial and start at 1, so this matches no records
		noEmployee := 0
		return &noEmployee
	}
}

func (c Checker) EditAll(resource Resource) bool {
	return c.grants.Edit[resource] == All
}

// A record is editable only when every one of its employees is editable
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

// Whether the actor can manage team members with the given role, including assigning it
func (c Checker) Manage(role types.EmployeeRole) bool {
	return slices.Contains(c.grants.Manage, role)
}
