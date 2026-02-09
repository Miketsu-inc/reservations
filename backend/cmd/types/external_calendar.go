package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type EventSource struct {
	source string
}

func (e EventSource) String() string {
	return e.source
}

var (
	EventSourceInternal = EventSource{"internal"}
	EventSourceGoogle   = EventSource{"google"}
)

func NewEventSource(sourceStr string) (EventSource, error) {
	switch strings.ToLower(sourceStr) {
	case "internal":
		return EventSourceInternal, nil
	case "google":
		return EventSourceGoogle, nil
	default:
		return EventSource{}, fmt.Errorf("invalid event source: %s", sourceStr)
	}
}

func (e EventSource) Value() (driver.Value, error) {
	return e.source, nil
}

func (e *EventSource) Scan(src any) error {
	sourceStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(sourceStr) == 0 {
		return nil
	}

	source, err := NewEventSource(sourceStr)
	if err != nil {
		return err
	}

	*e = source
	return nil
}

func (e EventSource) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.source)
}

func (e *EventSource) UnmarshalJSON(data []byte) error {
	var sourceStr string
	if err := json.Unmarshal(data, &sourceStr); err != nil {
		return err
	}

	source, err := NewEventSource(sourceStr)
	if err != nil {
		return err
	}

	*e = source
	return nil
}

type EventInternalType struct {
	itype string
}

func (e EventInternalType) String() string {
	return e.itype
}

var (
	EventInternalTypeBooking     = EventInternalType{"booking"}
	EventInternalTypeBlockedTime = EventInternalType{"blocked_time"}
)

func NewEventInternalType(typeStr string) (EventInternalType, error) {
	switch strings.ToLower(typeStr) {
	case "booking":
		return EventInternalTypeBooking, nil
	case "blocked_time":
		return EventInternalTypeBlockedTime, nil
	default:
		return EventInternalType{}, fmt.Errorf("invalid internal event type: %s", typeStr)
	}
}

func (e EventInternalType) Value() (driver.Value, error) {
	return e.itype, nil
}

func (e *EventInternalType) Scan(src any) error {
	typeStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(typeStr) == 0 {
		return nil
	}

	itype, err := NewEventInternalType(typeStr)
	if err != nil {
		return err
	}

	*e = itype
	return nil
}

func (e EventInternalType) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.itype)
}

func (e *EventInternalType) UnmarshalJSON(data []byte) error {
	var typeStr string
	if err := json.Unmarshal(data, &typeStr); err != nil {
		return err
	}

	itype, err := NewEventInternalType(typeStr)
	if err != nil {
		return err
	}

	*e = itype
	return nil
}
