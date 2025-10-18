package employee

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type Role struct {
	role string
}

func (r Role) String() string {
	return r.role
}

var (
	Staff = Role{"staff"}
	Admin = Role{"admin"}
	Owner = Role{"owner"}
)

func NewRole(roleStr string) (Role, error) {
	switch strings.ToLower(roleStr) {
	case "staff":
		return Staff, nil
	case "admin":
		return Admin, nil
	case "owner":
		return Owner, nil
	default:
		return Role{}, fmt.Errorf("invalid role: %s", roleStr)
	}
}

func (r Role) Value() (driver.Value, error) {
	return r.role, nil
}

func (r *Role) Scan(src any) error {
	roleStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(roleStr) == 0 {
		return nil
	}

	role, err := NewRole(roleStr)
	if err != nil {
		return err
	}

	*r = role
	return nil
}
