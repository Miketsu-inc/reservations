package subscription

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type Tier struct {
	tier string
}

func (t Tier) String() string {
	return t.tier
}

var (
	Free       = Tier{"free"}
	Pro        = Tier{"pro"}
	Enterprise = Tier{"enterprise"}
)

func NewTier(tierStr string) (Tier, error) {
	switch strings.ToLower(tierStr) {
	case "free":
		return Free, nil
	case "pro":
		return Pro, nil
	case "enterprise":
		return Enterprise, nil
	default:
		return Tier{}, fmt.Errorf("invalid tier: %s", tierStr)
	}
}

func (t Tier) Value() (driver.Value, error) {
	return t.tier, nil
}

func (t *Tier) Scan(src any) error {
	tierStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(tierStr) == 0 {
		return nil
	}

	tier, err := NewTier(tierStr)
	if err != nil {
		return err
	}

	*t = tier
	return nil
}
