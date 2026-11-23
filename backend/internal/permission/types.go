package permission

import (
	"encoding/json"

	"github.com/miketsu-inc/reservations/backend/internal/types"
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
	DeleteMerchant     = Action{"merchant:delete"}
	RenameMerchant     = Action{"merchant:rename"}
	EditSettings       = Action{"settings:edit"}
	EditLocationImages = Action{"location_images:edit"}
	ManageTeam         = Action{"team:manage"}
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
