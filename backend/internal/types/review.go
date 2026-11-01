package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

type ReviewStatus struct {
	status string
}

func (s ReviewStatus) String() string {
	return s.status
}

var (
	ReviewStatusPublished = ReviewStatus{"published"}
	ReviewStatusReported  = ReviewStatus{"reported"}
	ReviewStatusRemoved   = ReviewStatus{"removed"}
)

func NewReviewStatus(statusStr string) (ReviewStatus, error) {
	switch strings.ToLower(statusStr) {
	case "published":
		return ReviewStatusPublished, nil
	case "reported":
		return ReviewStatusReported, nil
	case "removed":
		return ReviewStatusRemoved, nil
	default:
		return ReviewStatus{}, fmt.Errorf("invalid review status: %s", statusStr)
	}
}

func (s ReviewStatus) Value() (driver.Value, error) {
	return s.status, nil
}

func (s *ReviewStatus) Scan(src any) error {
	statusStr, ok := src.(string)
	if !ok {
		return fmt.Errorf("value is not a string: %v", src)
	}

	if len(statusStr) == 0 {
		return nil
	}

	status, err := NewReviewStatus(statusStr)
	if err != nil {
		return nil
	}

	*s = status

	return nil
}

func (s ReviewStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.status)
}
