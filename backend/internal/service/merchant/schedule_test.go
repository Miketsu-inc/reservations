package merchant

import (
	"fmt"
	"testing"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ct(year int, month time.Month, day int, timeStr string, loc *time.Location) time.Time {
	t, _ := time.Parse("15:04", timeStr)
	return time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, loc)
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func ctReserved(year int, month time.Month, day int, start, end string, loc *time.Location) domain.BookingSlot {
	return domain.BookingSlot{
		FromDate: ct(year, month, day, start, loc).UTC(),
		ToDate:   ct(year, month, day, end, loc).UTC(),
	}
}

func ctBH(timeStr string) time.Time {
	t, _ := time.Parse("15:04", timeStr)
	return time.Date(0, time.January, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)
}

func formatTimes(times []time.Time) []string {
	formatted := make([]string, len(times))
	for i, t := range times {
		formatted[i] = fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
	}
	return formatted
}

func TestHasAllDayBlock(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Budapest")

	year := 2025
	month := time.July
	day := 1

	t.Run("No blocked times", func(t *testing.T) {
		assert.False(t, hasAllDayBlock([]domain.BlockedTimes{}))
	})

	t.Run("Only partial day blocks", func(t *testing.T) {
		blocked := []domain.BlockedTimes{
			{
				IsAllDay: false,
				FromDate: timePtr(ct(year, month, day-1, "10:00", tz)),
				ToDate:   timePtr(ct(year, month, day, "11:00", tz)),
			},
			{
				IsAllDay: false,
				FromDate: timePtr(ct(year, month, day, "15:00", tz)),
				ToDate:   timePtr(ct(year, month, day, "17:00", tz)),
			},
		}
		assert.False(t, hasAllDayBlock(blocked))
	})

	t.Run("Contains all day block", func(t *testing.T) {

		blocked := []domain.BlockedTimes{
			{
				IsAllDay: false,
				FromDate: timePtr(ct(year, month, day-1, "10:00", tz)),
				ToDate:   timePtr(ct(year, month, day, "11:00", tz)),
			},
			{
				IsAllDay:   true,
				BlockedDay: timePtr(time.Date(year, month, day, 0, 0, 0, 0, time.UTC)),
			},
		}
		assert.True(t, hasAllDayBlock(blocked))
	})
}

func TestFilterBlockedTimesForDayUsesBlockedDay(t *testing.T) {
	tz, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	matchingDay := time.Date(2026, time.September, 24, 0, 0, 0, 0, time.UTC)
	otherDay := matchingDay.AddDate(0, 0, 1)
	blocks := []domain.BlockedTimes{
		{IsAllDay: true, BlockedDay: &matchingDay},
		{IsAllDay: true, BlockedDay: &otherDay},
	}

	day := time.Date(2026, time.September, 24, 12, 0, 0, 0, tz)
	filtered := filterBlockedTimesForDay(blocks, day, tz)

	require.Len(t, filtered, 1)
	assert.Equal(t, matchingDay, *filtered[0].BlockedDay)
}

func TestHasNoPhaseConflict(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Budapest")
	year := 2026
	month := time.September
	day := 12
	bookingTime := ct(year, month, day, "10:00", tz)
	bufferZero := time.Duration(0)

	t.Run("1 active phase conflicts with reserved time", func(t *testing.T) {
		phases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}

		reserved := []domain.BookingSlot{
			ctReserved(year, month, day, "10:15", "10:45", tz),
		}

		blocked := []domain.BlockedTimes{}

		assert.False(t, hasNoPhaseConflict(bookingTime, phases, blocked, reserved, bufferZero, tz))
	})

	t.Run("1 active phase conflict with blocked time", func(t *testing.T) {
		phases := []domain.ServicePhase{{PhaseType: types.ServicePhaseTypeActive, Duration: 30}}

		reserved := []domain.BookingSlot{}

		blocked := []domain.BlockedTimes{{
			IsAllDay: false,
			FromDate: timePtr(ct(year, month, day, "06:45", tz)),
			ToDate:   timePtr(ct(year, month, day, "10:20", tz)),
		}}

		assert.False(t, hasNoPhaseConflict(bookingTime, phases, blocked, reserved, bufferZero, tz))
	})

	t.Run("Multiple phases with wait in the start", func(t *testing.T) {
		phases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},
			{PhaseType: types.ServicePhaseTypeActive, Duration: 45},
		}

		reserved := []domain.BookingSlot{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "9:15", "10:00", tz),
		}

		blocked := []domain.BlockedTimes{}

		assert.True(t, hasNoPhaseConflict(bookingTime, phases, blocked, reserved, bufferZero, tz))
	})

	t.Run("Multiple phases with wait in the middle", func(t *testing.T) {
		phases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 15},
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},
			{PhaseType: types.ServicePhaseTypeActive, Duration: 45},
		}

		reserved := []domain.BookingSlot{
			ctReserved(year, month, day, "10:15", "10:45", tz),
			ctReserved(year, month, day, "11:30", "12:00", tz),
		}

		blocked := []domain.BlockedTimes{}

		assert.True(t, hasNoPhaseConflict(bookingTime, phases, blocked, reserved, bufferZero, tz))
	})

	t.Run("Multiple phases with wait in middle conflict", func(t *testing.T) {
		phases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 15},
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},
			{PhaseType: types.ServicePhaseTypeActive, Duration: 45},
		}

		reserved := []domain.BookingSlot{
			ctReserved(year, month, day, "10:30", "11:00", tz),
		}

		blocked := []domain.BlockedTimes{}

		assert.False(t, hasNoPhaseConflict(bookingTime, phases, blocked, reserved, bufferZero, tz))
	})

	t.Run("Conflict with buffer time between bookings", func(t *testing.T) {
		phases := []domain.ServicePhase{{PhaseType: types.ServicePhaseTypeActive, Duration: 30}}
		buffer15 := 15 * time.Minute

		// with 15 min buffer closed period is 10:15 - 11:15
		reserved := []domain.BookingSlot{ctReserved(year, month, day, "10:30", "11:00", tz)}

		blocked := []domain.BlockedTimes{}

		assert.False(t, hasNoPhaseConflict(bookingTime, phases, blocked, reserved, buffer15, tz))
	})

	t.Run("Blocked time overlaps only a wait phase", func(t *testing.T) {
		phases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},
			{PhaseType: types.ServicePhaseTypeActive, Duration: 15},
		}

		reserved := []domain.BookingSlot{}

		blocked := []domain.BlockedTimes{{
			IsAllDay: false,
			FromDate: timePtr(ct(year, month, day, "10:30", tz)),
			ToDate:   timePtr(ct(year, month, day, "10:45", tz)),
		}}

		assert.True(t, hasNoPhaseConflict(bookingTime, phases, blocked, reserved, bufferZero, tz))
	})
}

func TestCalculateAvailableTimes(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Budapest")

	year := 2025
	month := time.July
	day := 1

	t.Run("Business hours", func(t *testing.T) {
		reserved := []domain.BookingSlot{}

		servicePhases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}
		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 0

		bookingDay := ct(2025, time.July, 1, "00:00", tz)

		businessHours := []domain.TimeSlot{
			{StartTime: ctBH("09:30"), EndTime: ctBH("11:30")},
			{StartTime: ctBH("13:00"), EndTime: ctBH("16:15")},
		}

		expected := []time.Time{
			ct(year, month, day, "09:30", tz), ct(year, month, day, "09:45", tz),
			ct(year, month, day, "10:00", tz), ct(year, month, day, "10:15", tz),
			ct(year, month, day, "10:30", tz), ct(year, month, day, "10:45", tz),
			ct(year, month, day, "11:00", tz),
			ct(year, month, day, "13:00", tz), ct(year, month, day, "13:15", tz),
			ct(year, month, day, "13:30", tz), ct(year, month, day, "13:45", tz),
			ct(year, month, day, "14:00", tz), ct(year, month, day, "14:15", tz),
			ct(year, month, day, "14:30", tz), ct(year, month, day, "14:45", tz),
			ct(year, month, day, "15:00", tz), ct(year, month, day, "15:15", tz),
			ct(year, month, day, "15:30", tz), ct(year, month, day, "15:45", tz),
		}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		blocked := []domain.BlockedTimes{}

		result := CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, expected, result, "Available times do not match")
	})

	t.Run("Close current time", func(t *testing.T) {
		reserved := []domain.BookingSlot{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:00", "11:45", tz),
			ctReserved(year, month, day, "13:00", "14:00", tz),
		}

		servicePhases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 15},
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},
			{PhaseType: types.ServicePhaseTypeActive, Duration: 45},
		}
		serviceDuration := 90
		bookingWindowMin, bufferTime := 0, 0

		bookingDay := ct(year, month, day, "00:00", tz)

		businessHours := []domain.TimeSlot{
			{StartTime: ctBH("09:00"), EndTime: ctBH("16:00")},
		}

		expected := []time.Time{
			ct(year, month, day, "14:30", tz),
		}

		blocked := []domain.BlockedTimes{}

		currentTime := ct(2025, time.July, 1, "14:20", tz)

		result := CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, expected, result, "Available times do not match")
	})
}

func TestCalculateAvailableTimesPeriod(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Budapest")

	t.Run("Empty period", func(t *testing.T) {
		startDate := ct(2025, time.July, 1, "00:00", tz)
		endDate := ct(2025, time.June, 30, "23:59", tz)

		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 0

		blocked := []domain.BlockedTimes{}

		reserved := []domain.BookingSlot{}

		servicePhases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}

		businessHours := domain.BusinessHours{
			3: {
				{StartTime: ctBH("09:00"), EndTime: ctBH("11:00")},
			},
		}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		results := CalculateAvailableTimesPeriod(
			reserved,
			blocked,
			servicePhases,
			serviceDuration,
			bufferTime,
			bookingWindowMin,
			startDate, endDate,
			businessHours,
			currentTime,
			tz,
		)

		assert.Equal(t, 0, len(results), "Should return empty results for invalid date range")
	})

	t.Run("Empty Business Hours", func(t *testing.T) {
		startDate := ct(2025, time.July, 1, "00:00", tz)
		endDate := ct(2025, time.July, 3, "23:59", tz)

		reserved := []domain.BookingSlot{
			ctReserved(2025, time.July, 2, "10:00", "10:30", tz),
		}

		servicePhases := []domain.ServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}
		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 0

		businessHours := domain.BusinessHours{
			2: {}, // Tuesday (July 1, 2025)
			3: { // Wednesday (July 2, 2025)
				{StartTime: ctBH("09:00"), EndTime: ctBH("11:00")},
			},
			4: { // Thursday (July 3, 2025)
				{StartTime: ctBH("09:00"), EndTime: ctBH("11:00")},
			},
		}

		blocked := []domain.BlockedTimes{}

		currentTime := ct(2025, time.June, 12, "00:00", tz)

		results := CalculateAvailableTimesPeriod(
			reserved,
			blocked,
			servicePhases,
			serviceDuration,
			bufferTime,
			bookingWindowMin,
			startDate,
			endDate,
			businessHours,
			currentTime,
			tz,
		)

		assert.Equal(t, 2, len(results), "Expected 2 days of results")

		expectedDay2 := formatTimes([]time.Time{
			ct(2025, time.July, 2, "09:00", tz),
			ct(2025, time.July, 2, "09:15", tz),
			ct(2025, time.July, 2, "09:30", tz),
			ct(2025, time.July, 2, "10:30", tz),
		})

		expectedDay3 := formatTimes([]time.Time{
			ct(2025, time.July, 3, "09:00", tz),
			ct(2025, time.July, 3, "09:15", tz),
			ct(2025, time.July, 3, "09:30", tz),
			ct(2025, time.July, 3, "09:45", tz),
			ct(2025, time.July, 3, "10:00", tz),
			ct(2025, time.July, 3, "10:15", tz),
			ct(2025, time.July, 3, "10:30", tz),
		})

		assert.ElementsMatch(t, expectedDay2, append(results[0].Morning, results[0].Afternoon...), "Day 2 times mismatch")
		assert.ElementsMatch(t, expectedDay3, append(results[1].Morning, results[1].Afternoon...), "Day 3 times mismatch")
	})
}

func TestCacluateAvailableDays(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Budapest")

	year := 2025
	month := time.July
	day := 1

	startDate := ct(year, month, day, "00:00", tz) // Tuesday
	endDate := ct(year, month, day+2, "00:00", tz)

	servicePhases := []domain.ServicePhase{{PhaseType: types.ServicePhaseTypeActive, Duration: 60}}

	businessHours := domain.BusinessHours{
		2: {},                                                   // Tuesday: closed
		3: {{StartTime: ctBH("09:00"), EndTime: ctBH("10:00")}}, // Wednesday
		4: {{StartTime: ctBH("09:00"), EndTime: ctBH("10:00")}}, // Thursday
	}

	reserved := []domain.BookingSlot{
		ctReserved(2025, time.July, 2, "09:00", "10:00", tz),
	}

	blocked := []domain.BlockedTimes{}

	currentTime := ct(2025, time.June, 1, "00:00", tz)

	result := CalculateAvailableDays(reserved, blocked, servicePhases, 60, 0, 0, startDate, endDate, businessHours, currentTime, tz)

	assert.Len(t, result, 3, "There should be only 3 day in the result")

	assert.Equal(t, "2025-07-01", result[0].Date)
	assert.False(t, result[0].IsAvailable, "Closed day should not be available")

	assert.Equal(t, "2025-07-02", result[1].Date)
	assert.False(t, result[1].IsAvailable, "Open day no empty slots should not be available")

	assert.Equal(t, "2025-07-03", result[2].Date)
	assert.True(t, result[2].IsAvailable, "Open day with open slots should be available")
}

func TestIsValidBookingSlot(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Budapest")

	year := 2025
	month := time.July
	day := 1

	businessHours := []domain.TimeSlot{
		{StartTime: ctBH("09:00"), EndTime: ctBH("17:00")},
	}
	servicePhases := []domain.ServicePhase{
		{PhaseType: types.ServicePhaseTypeActive, Duration: 60},
	}

	reserved := []domain.BookingSlot{}

	blocked := []domain.BlockedTimes{}

	currentTime := ct(year, month, day, "00:00", tz)

	bookingWindowMin, bufferTime, bookingWindowMax := 0, 0, 1

	t.Run("Valid within business hours", func(t *testing.T) {
		slot := domain.TimeSlot{
			StartTime: ct(year, month, day, "10:00", tz),
			EndTime:   ct(year, month, day, "11:00", tz),
		}

		assert.True(t, IsValidBookingSlot(slot, reserved, blocked, servicePhases, businessHours, bufferTime, bookingWindowMin, bookingWindowMax, currentTime, tz))
	})

	t.Run("Outside business hours", func(t *testing.T) {
		slot := domain.TimeSlot{
			StartTime: ct(year, month, day, "16:30", tz),
			EndTime:   ct(year, month, day, "17:30", tz),
		}
		assert.False(t, IsValidBookingSlot(slot, reserved, blocked, servicePhases, businessHours, 0, 0, 1, currentTime, tz))
	})

	t.Run("Exceeds bookingWindow", func(t *testing.T) {
		slot := domain.TimeSlot{
			StartTime: ct(year+1, month, 1, "10:00", tz),
			EndTime:   ct(year+1, month, 1, "11:00", tz),
		}

		assert.False(t, IsValidBookingSlot(slot, reserved, blocked, servicePhases, businessHours, 0, 0, 1, currentTime, tz))
	})

}
