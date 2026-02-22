package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type ServicePhaseType struct {
	ptype string
}

func (s ServicePhaseType) String() string {
	return s.ptype
}

var (
	ServicePhaseTypeActive = ServicePhaseType{"active"}
	ServicePhaseTypeWait   = ServicePhaseType{"wait"}
)

func NewServicePhaseType(typeStr string) (ServicePhaseType, error) {
	switch strings.ToLower(typeStr) {
	case "active":
		return ServicePhaseTypeActive, nil
	case "wait":
		return ServicePhaseTypeWait, nil
	default:
		return ServicePhaseType{}, fmt.Errorf("invalid ServicePhaseType: %s", typeStr)

	}
}

func (t ServicePhaseType) Value() (driver.Value, error) {
	return t.ptype, nil
}

func (t *ServicePhaseType) Scan(src any) error {
	typeStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(typeStr) == 0 {
		return nil
	}

	ptype, err := NewServicePhaseType(typeStr)
	if err != nil {
		return err
	}

	*t = ptype
	return nil
}

func (t ServicePhaseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.ptype)
}

func (t *ServicePhaseType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	ptype, err := NewServicePhaseType(s)
	if err != nil {
		return err
	}

	*t = ptype
	return nil
}

type PriceType struct {
	ptype string
}

func (p PriceType) String() string {
	return p.ptype
}

var (
	PriceTypeFree  = PriceType{"free"}
	PriceTypeFrom  = PriceType{"from"}
	PriceTypeFixed = PriceType{"fixed"}
)

func NewPriceType(typeStr string) (PriceType, error) {
	switch strings.ToLower(typeStr) {
	case "fixed":
		return PriceTypeFixed, nil
	case "from":
		return PriceTypeFrom, nil
	case "free":
		return PriceTypeFree, nil
	default:
		return PriceType{}, fmt.Errorf("invalid Pricetype: %s", typeStr)
	}
}

func (p PriceType) Value() (driver.Value, error) {
	return p.ptype, nil
}

func (t *PriceType) Scan(src any) error {
	typeStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(typeStr) == 0 {
		return nil
	}

	ptype, err := NewPriceType(typeStr)
	if err != nil {
		return err
	}

	*t = ptype
	return nil
}

func (t PriceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.ptype)
}

func (t *PriceType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	ptype, err := NewPriceType(s)
	if err != nil {
		return err
	}

	*t = ptype
	return nil
}
