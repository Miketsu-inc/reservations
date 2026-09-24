package blockedtimes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapToNewInputAllDay(t *testing.T) {
	input, err := mapToNewInput(newReq{
		Name:       "Day off",
		BlockedDay: "2026-09-24",
		IsAllDay:   true,
	})

	require.NoError(t, err)
	require.NotNil(t, input.BlockedDay)
	assert.Equal(t, "2026-09-24", input.BlockedDay.Format(time.DateOnly))
	assert.Nil(t, input.FromDate)
	assert.Nil(t, input.ToDate)
}

func TestMapToNewInputTimed(t *testing.T) {
	input, err := mapToNewInput(newReq{
		Name:     "Appointment",
		FromDate: "2026-09-24T09:00:00+02:00",
		ToDate:   "2026-09-24T10:00:00+02:00",
	})

	require.NoError(t, err)
	assert.Nil(t, input.BlockedDay)
	require.NotNil(t, input.FromDate)
	require.NotNil(t, input.ToDate)
	assert.True(t, input.ToDate.After(*input.FromDate))
}

func TestMapToNewInputRejectsInvalidBlockedDay(t *testing.T) {
	_, err := mapToNewInput(newReq{
		Name:       "Day off",
		BlockedDay: "2026-02-30",
		IsAllDay:   true,
	})

	assert.Error(t, err)
}
