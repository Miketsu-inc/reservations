package booking

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type Status struct {
	status string
}

func (s Status) String() string {
	return s.status
}

var (
	Booked    = Status{"booked"}
	Confirmed = Status{"confirmed"}
	Completed = Status{"completed"}
	Cancelled = Status{"cancelled"}
	NoShow    = Status{"no-show"}
)

func NewStatus(statusStr string) (Status, error) {
	switch strings.ToLower(statusStr) {
	case "booked":
		return Booked, nil
	case "confirmed":
		return Confirmed, nil
	case "completed":
		return Completed, nil
	case "cancelled":
		return Cancelled, nil
	case "no-show":
		return NoShow, nil
	default:
		return Status{}, fmt.Errorf("invalid booking Status: %s", statusStr)
	}
}

func (s Status) Value() (driver.Value, error) {
	return s.status, nil
}

func (s *Status) Scan(src any) error {
	statusStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(statusStr) == 0 {
		return nil
	}

	status, err := NewStatus(statusStr)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.status)
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var statusStr string
	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}

	status, err := NewStatus(statusStr)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

type Type struct {
	btype string
}

func (t Type) String() string {
	return t.btype
}

var (
	Appointment = Type{"appointment"}
	Event       = Type{"event"}
	Class       = Type{"class"}
)

func NewType(typeStr string) (Type, error) {
	switch strings.ToLower(typeStr) {
	case "appointment":
		return Appointment, nil
	case "event":
		return Event, nil
	case "class":
		return Class, nil
	default:
		return Type{}, fmt.Errorf("invalid booking Type: %s", typeStr)
	}
}

func (t Type) Value() (driver.Value, error) {
	return t.btype, nil
}

func (t *Type) Scan(src any) error {
	typeStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(typeStr) == 0 {
		return nil
	}

	btype, err := NewType(typeStr)
	if err != nil {
		return err
	}

	*t = btype
	return nil
}

func (t Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.btype)
}

func (t *Type) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	btype, err := NewType(s)
	if err != nil {
		return err
	}

	*t = btype
	return nil
}
