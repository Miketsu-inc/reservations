package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type BookingStatus struct {
	status string
}

func (s BookingStatus) String() string {
	return s.status
}

var (
	BookingStatusBooked    = BookingStatus{"booked"}
	BookingStatusConfirmed = BookingStatus{"confirmed"}
	BookingStatusCompleted = BookingStatus{"completed"}
	BookingStatusCancelled = BookingStatus{"cancelled"}
	BookingStatusNoShow    = BookingStatus{"no-show"}
)

func NewBookingStatus(statusStr string) (BookingStatus, error) {
	switch strings.ToLower(statusStr) {
	case "booked":
		return BookingStatusBooked, nil
	case "confirmed":
		return BookingStatusConfirmed, nil
	case "completed":
		return BookingStatusCompleted, nil
	case "cancelled":
		return BookingStatusCancelled, nil
	case "no-show":
		return BookingStatusNoShow, nil
	default:
		return BookingStatus{}, fmt.Errorf("invalid booking Status: %s", statusStr)
	}
}

func (s BookingStatus) Value() (driver.Value, error) {
	return s.status, nil
}

func (s *BookingStatus) Scan(src any) error {
	statusStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(statusStr) == 0 {
		return nil
	}

	status, err := NewBookingStatus(statusStr)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

func (s BookingStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.status)
}

func (s *BookingStatus) UnmarshalJSON(data []byte) error {
	var statusStr string
	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}

	status, err := NewBookingStatus(statusStr)
	if err != nil {
		return err
	}

	*s = status
	return nil
}

type BookingType struct {
	btype string
}

func (t BookingType) String() string {
	return t.btype
}

var (
	BookingTypeAppointment = BookingType{"appointment"}
	BookingTypeEvent       = BookingType{"event"}
	BookingTypeClass       = BookingType{"class"}
)

func NewBookingType(typeStr string) (BookingType, error) {
	switch strings.ToLower(typeStr) {
	case "appointment":
		return BookingTypeAppointment, nil
	case "event":
		return BookingTypeEvent, nil
	case "class":
		return BookingTypeClass, nil
	default:
		return BookingType{}, fmt.Errorf("invalid booking Type: %s", typeStr)
	}
}

func (t BookingType) Value() (driver.Value, error) {
	return t.btype, nil
}

func (t *BookingType) Scan(src any) error {
	typeStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(typeStr) == 0 {
		return nil
	}

	btype, err := NewBookingType(typeStr)
	if err != nil {
		return err
	}

	*t = btype
	return nil
}

func (t BookingType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.btype)
}

func (t *BookingType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	btype, err := NewBookingType(s)
	if err != nil {
		return err
	}

	*t = btype
	return nil
}
