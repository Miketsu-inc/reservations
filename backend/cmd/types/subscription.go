package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type SubTier struct {
	tier string
}

func (t SubTier) String() string {
	return t.tier
}

var (
	SubTierFree       = SubTier{"free"}
	SubTierPro        = SubTier{"pro"}
	SubTierEnterprise = SubTier{"enterprise"}
)

func NewSubTier(tierStr string) (SubTier, error) {
	switch strings.ToLower(tierStr) {
	case "free":
		return SubTierFree, nil
	case "pro":
		return SubTierPro, nil
	case "enterprise":
		return SubTierEnterprise, nil
	default:
		return SubTier{}, fmt.Errorf("invalid tier: %s", tierStr)
	}
}

func (t SubTier) Value() (driver.Value, error) {
	return t.tier, nil
}

func (t *SubTier) Scan(src any) error {
	tierStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(tierStr) == 0 {
		return nil
	}

	tier, err := NewSubTier(tierStr)
	if err != nil {
		return err
	}

	*t = tier
	return nil
}

func (t SubTier) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.tier)
}

func (t *SubTier) UnmarshalJSON(data []byte) error {
	var tierStr string
	if err := json.Unmarshal(data, &tierStr); err != nil {
		return err
	}

	tier, err := NewSubTier(tierStr)
	if err != nil {
		return err
	}

	*t = tier
	return nil
}
