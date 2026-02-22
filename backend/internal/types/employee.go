package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type EmployeeRole struct {
	role string
}

func (r EmployeeRole) String() string {
	return r.role
}

var (
	EmployeeRoleStaff = EmployeeRole{"staff"}
	EmployeeRoleAdmin = EmployeeRole{"admin"}
	EmployeeRoleOwner = EmployeeRole{"owner"}
)

func NewEmployeeRole(roleStr string) (EmployeeRole, error) {
	switch strings.ToLower(roleStr) {
	case "staff":
		return EmployeeRoleStaff, nil
	case "admin":
		return EmployeeRoleAdmin, nil
	case "owner":
		return EmployeeRoleOwner, nil
	default:
		return EmployeeRole{}, fmt.Errorf("invalid role: %s", roleStr)
	}
}

func (r EmployeeRole) Value() (driver.Value, error) {
	return r.role, nil
}

func (r *EmployeeRole) Scan(src any) error {
	roleStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(roleStr) == 0 {
		return nil
	}

	role, err := NewEmployeeRole(roleStr)
	if err != nil {
		return err
	}

	*r = role
	return nil
}

func (r EmployeeRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.role)
}

func (r *EmployeeRole) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	role, err := NewEmployeeRole(s)
	if err != nil {
		return err
	}

	*r = role
	return nil
}
