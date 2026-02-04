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
	EventSourceGoogle = EventSource{"google"}
)

func NewEventSource(sourceStr string) (EventSource, error) {
	switch strings.ToLower(sourceStr) {
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
