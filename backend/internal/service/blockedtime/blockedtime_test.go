package blockedtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateDateShape(t *testing.T) {
	day := time.Date(2026, time.September, 24, 0, 0, 0, 0, time.UTC)
	from := time.Date(2026, time.September, 24, 9, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)

	tests := []struct {
		name       string
		isAllDay   bool
		blockedDay *time.Time
		fromDate   *time.Time
		toDate     *time.Time
		wantError  bool
	}{
		{name: "all day", isAllDay: true, blockedDay: &day},
		{name: "timed", fromDate: &from, toDate: &to},
		{name: "all day without day", isAllDay: true, wantError: true},
		{name: "all day with timestamps", isAllDay: true, blockedDay: &day, fromDate: &from, toDate: &to, wantError: true},
		{name: "timed without timestamps", wantError: true},
		{name: "timed with day", blockedDay: &day, fromDate: &from, toDate: &to, wantError: true},
		{name: "timed reversed", fromDate: &to, toDate: &from, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDateShape(test.isAllDay, test.blockedDay, test.fromDate, test.toDate)
			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
