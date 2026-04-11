package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type ApprovalType struct {
	atype string
}

func (t ApprovalType) String() string {
	return t.atype
}

var (
	ApprovalTypeAuto         = ApprovalType{"auto"}
	ApprovalTypeManual       = ApprovalType{"manual"}
	ApprovalTypeManualForNew = ApprovalType{"manual_for_new"}
)

func NewApprovalPolicy(typeStr string) (ApprovalType, error) {
	switch strings.ToLower(typeStr) {
	case "auto":
		return ApprovalTypeAuto, nil
	case "manual":
		return ApprovalTypeManual, nil
	case "manual_for_new":
		return ApprovalTypeManualForNew, nil
	default:
		return ApprovalType{}, fmt.Errorf("invalid approval policy: %s", typeStr)
	}
}

func (t ApprovalType) Value() (driver.Value, error) {
	return t.atype, nil
}

func (t *ApprovalType) Scan(src any) error {
	typeStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("values is not a string: %v", src)
	}

	if len(typeStr) == 0 {
		return nil
	}

	atype, err := NewApprovalPolicy(typeStr)
	if err != nil {
		return err
	}

	*t = atype
	return nil
}

func (t ApprovalType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.atype)
}

func (t *ApprovalType) UnmarshalJSON(data []byte) error {
	var typeStr string
	if err := json.Unmarshal(data, &typeStr); err != nil {
		return err
	}

	atype, err := NewApprovalPolicy(typeStr)
	if err != nil {
		return err
	}

	*t = atype
	return nil
}
