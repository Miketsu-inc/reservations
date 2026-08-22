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

type EmployeeInvitationRole struct {
	role string
}

func (r EmployeeInvitationRole) String() string {
	return r.role
}

var (
	EmployeeInvitationRoleStaff = EmployeeInvitationRole{"staff"}
	EmployeeInvitationRoleAdmin = EmployeeInvitationRole{"admin"}
)

func NewEmployeeInvitationRole(roleStr string) (EmployeeInvitationRole, error) {
	switch strings.ToLower(roleStr) {
	case "staff":
		return EmployeeInvitationRoleStaff, nil
	case "admin":
		return EmployeeInvitationRoleAdmin, nil
	default:
		return EmployeeInvitationRole{}, fmt.Errorf("invalid role: %s", roleStr)
	}
}

func (r EmployeeInvitationRole) ToEmployeeRole() EmployeeRole {
	switch r {
	case EmployeeInvitationRoleStaff:
		return EmployeeRoleStaff
	case EmployeeInvitationRoleAdmin:
		return EmployeeRoleAdmin
	default:
		return EmployeeRoleStaff
	}
}

func (r EmployeeInvitationRole) Value() (driver.Value, error) {
	return r.role, nil
}

func (r *EmployeeInvitationRole) Scan(src any) error {
	roleStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(roleStr) == 0 {
		return nil
	}

	role, err := NewEmployeeInvitationRole(roleStr)
	if err != nil {
		return err
	}

	*r = role
	return nil
}

func (r EmployeeInvitationRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.role)
}

func (r *EmployeeInvitationRole) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	role, err := NewEmployeeInvitationRole(s)
	if err != nil {
		return err
	}

	*r = role
	return nil
}

type EmployeeInvitationStatus struct {
	status string
}

func (s EmployeeInvitationStatus) String() string {
	return s.status
}

var (
	EmployeeInvitationStatusPending  = EmployeeInvitationStatus{"pending"}
	EmployeeInvitationStatusAccepted = EmployeeInvitationStatus{"accepted"}
	EmployeeInvitationStatusRevoked  = EmployeeInvitationStatus{"revoked"}
	EmployeeInvitationStatusDeclined = EmployeeInvitationStatus{"declined"}
	EmployeeInvitationStatusExpired  = EmployeeInvitationStatus{"expired"}
)

func NewEmployeeInvitationStatus(statusStr string) (EmployeeInvitationStatus, error) {
	switch strings.ToLower(statusStr) {
	case "pending":
		return EmployeeInvitationStatusPending, nil
	case "accepted":
		return EmployeeInvitationStatusAccepted, nil
	case "revoked":
		return EmployeeInvitationStatusRevoked, nil
	case "declined":
		return EmployeeInvitationStatusDeclined, nil
	case "expired":
		return EmployeeInvitationStatusExpired, nil
	default:
		return EmployeeInvitationStatus{}, fmt.Errorf("invalid status: %s", statusStr)
	}
}

func (s EmployeeInvitationStatus) Value() (driver.Value, error) {
	return s.status, nil
}

func (s *EmployeeInvitationStatus) Scan(src any) error {
	statusStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(statusStr) == 0 {
		return nil
	}

	status, err := NewEmployeeInvitationStatus(statusStr)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

func (s EmployeeInvitationStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.status)
}

func (s *EmployeeInvitationStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	status, err := NewEmployeeInvitationStatus(str)
	if err != nil {
		return err
	}

	*s = status
	return nil
}
