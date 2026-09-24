package externalcalendar

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/calendar/v3"
)

func TestEventToBlockedTimeAllDay(t *testing.T) {
	event := &calendar.Event{
		Id:      "event-id",
		Summary: "Day off",
		Start:   &calendar.EventDateTime{Date: "2026-09-24"},
		End:     &calendar.EventDateTime{Date: "2026-09-25"},
	}

	blockedTime, err := eventToBlockedTime(event, uuid.New(), time.UTC)

	require.NoError(t, err)
	assert.True(t, blockedTime.IsAllDay)
	require.NotNil(t, blockedTime.BlockedDay)
	assert.Equal(t, "2026-09-24", blockedTime.BlockedDay.Format(time.DateOnly))
	assert.Nil(t, blockedTime.FromDate)
	assert.Nil(t, blockedTime.ToDate)
}

func TestEventToBlockedTimeRejectsMultiDayEvent(t *testing.T) {
	event := &calendar.Event{
		Id:    "event-id",
		Start: &calendar.EventDateTime{Date: "2026-09-24"},
		End:   &calendar.EventDateTime{Date: "2026-09-26"},
	}

	_, err := eventToBlockedTime(event, uuid.New(), time.UTC)

	assert.Error(t, err)
}

func TestBlockedTimeToGoogleEventAllDay(t *testing.T) {
	day := time.Date(2026, time.September, 24, 0, 0, 0, 0, time.UTC)
	event := blockedTimeToGoogleEvent(domain.BlockedTime{
		Name:       "Day off",
		BlockedDay: &day,
		IsAllDay:   true,
	}, "Europe/Budapest")

	assert.Equal(t, "2026-09-24", event.Start.Date)
	assert.Equal(t, "2026-09-25", event.End.Date)
	assert.Empty(t, event.Start.DateTime)
	assert.Empty(t, event.End.DateTime)
}
