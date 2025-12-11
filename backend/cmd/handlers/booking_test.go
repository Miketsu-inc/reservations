package handlers_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/miketsu-inc/reservations/backend/cmd/handlers"
	"github.com/miketsu-inc/reservations/backend/cmd/types"
	"github.com/stretchr/testify/assert"
)

func ct(year int, month time.Month, day int, timeStr string, loc *time.Location) time.Time {
	t, _ := time.Parse("15:04", timeStr)
	return time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, loc)
}

func ctReserved(year int, month time.Month, day int, start, end string, loc *time.Location) database.BookingTime {
	return database.BookingTime{
		From_date: ct(year, month, day, start, loc).UTC(),
		To_date:   ct(year, month, day, end, loc).UTC(),
	}
}

func formatTimes(times []time.Time) []string {
	formatted := make([]string, len(times))
	for i, t := range times {
		formatted[i] = fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
	}
	return formatted
}

func TestCalculateAvailableTimes(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Budapest")

	year := 2025
	month := time.July
	day := 1

	t.Run("Business hours", func(t *testing.T) {
		reserved := []database.BookingTime{}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}
		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 0

		bookingDay := ct(2025, time.July, 1, "00:00", tz)

		businessHours := []database.TimeSlot{
			{StartTime: "09:30:00", EndTime: "11:30:00"},
			{StartTime: "13:00:00", EndTime: "16:15:00"},
		}

		expectedMorning := []string{"09:30", "09:45", "10:00", "10:15", "10:30", "10:45", "11:00"}
		expectedAfternoon := []string{"13:00", "13:15", "13:30", "13:45", "14:00", "14:15", "14:30", "14:45", "15:00", "15:15", "15:30", "15:45"}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		blocked := []database.BlockedTimes{}

		result := handlers.CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, expectedMorning, result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, expectedAfternoon, result.Afternoon, "Afternoon times do not match")
	})

	t.Run("One active phase", func(t *testing.T) {
		reserved := []database.BookingTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:00", "11:45", tz),
			ctReserved(year, month, day, "13:00", "14:00", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 60},
		}
		serviceDuration := 60
		bookingWindowMin, bufferTime := 0, 0

		bookingDay := ct(year, month, day, "00:00", tz)

		businessHours := []database.TimeSlot{
			{StartTime: "09:00:00", EndTime: "16:00:00"},
		}

		expectedMorning := []time.Time{
			ct(year, month, day, "09:00", tz),
			ct(year, month, day, "11:45", tz),
		}
		expectedAfternoon := []time.Time{
			ct(year, month, day, "12:00", tz),
			ct(year, month, day, "14:00", tz),
			ct(year, month, day, "14:15", tz),
			ct(year, month, day, "14:30", tz),
			ct(year, month, day, "14:45", tz),
			ct(year, month, day, "15:00", tz),
		}

		blocked := []database.BlockedTimes{}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		result := handlers.CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, formatTimes(expectedMorning), result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, formatTimes(expectedAfternoon), result.Afternoon, "Afternoon times do not match")
	})

	t.Run("Mutliple phases with wait at the start", func(t *testing.T) {
		reserved := []database.BookingTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:15", "11:30", tz),
			ctReserved(year, month, day, "13:00", "15:00", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},
			{PhaseType: types.ServicePhaseTypeActive, Duration: 15},
		}
		serviceDuration := 45
		bookingWindowMin, bufferTime := 0, 0

		bookingDay := ct(year, month, day, "00:00", tz)

		businessHours := []database.TimeSlot{
			{StartTime: "09:30:00", EndTime: "11:30:00"},
			{StartTime: "13:00:00", EndTime: "16:15:00"},
		}

		expectedMorning := []time.Time{
			ct(year, month, day, "10:00", tz),
			ct(year, month, day, "10:15", tz),
			ct(year, month, day, "10:30", tz),
		}
		expectedAfternoon := []time.Time{
			ct(year, month, day, "14:30", tz),
			ct(year, month, day, "14:45", tz),
			ct(year, month, day, "15:00", tz),
			ct(year, month, day, "15:15", tz),
			ct(year, month, day, "15:30", tz),
		}

		blocked := []database.BlockedTimes{}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		result := handlers.CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, formatTimes(expectedMorning), result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, formatTimes(expectedAfternoon), result.Afternoon, "Afternoon times do not match")
	})

	t.Run("Mutliple phases with wait in the middle", func(t *testing.T) {
		reserved := []database.BookingTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:15", "11:45", tz),
			ctReserved(year, month, day, "13:00", "14:00", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 15},
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},
			{PhaseType: types.ServicePhaseTypeActive, Duration: 45},
		}
		serviceDuration := 90
		bookingWindowMin, bufferTime := 0, 0

		bookingDay := ct(year, month, day, "00:00", tz)

		businessHours := []database.TimeSlot{
			{StartTime: "09:00:00", EndTime: "16:00:00"},
		}

		expectedMorning := []time.Time{
			ct(year, month, day, "09:45", tz),
			ct(year, month, day, "11:00", tz),
		}
		expectedAfternoon := []time.Time{
			ct(year, month, day, "14:00", tz),
			ct(year, month, day, "14:15", tz),
			ct(year, month, day, "14:30", tz),
		}

		blocked := []database.BlockedTimes{}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		result := handlers.CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, formatTimes(expectedMorning), result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, formatTimes(expectedAfternoon), result.Afternoon, "Afternoon times do not match")
	})

	t.Run("Close current time", func(t *testing.T) {
		reserved := []database.BookingTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:00", "11:45", tz),
			ctReserved(year, month, day, "13:00", "14:00", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 15},
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},
			{PhaseType: types.ServicePhaseTypeActive, Duration: 45},
		}
		serviceDuration := 90
		bookingWindowMin, bufferTime := 0, 0

		bookingDay := ct(year, month, day, "00:00", tz)

		businessHours := []database.TimeSlot{
			{StartTime: "09:00:00", EndTime: "16:00:00"},
		}

		expectedMorning := []time.Time{}
		expectedAfternoon := []time.Time{
			ct(year, month, day, "14:30", tz),
		}

		blocked := []database.BlockedTimes{}

		currentTime := ct(2025, time.July, 1, "14:20", tz)

		result := handlers.CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, formatTimes(expectedMorning), result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, formatTimes(expectedAfternoon), result.Afternoon, "Afternoon times do not match")
	})

	t.Run("Buffer time between bookings", func(t *testing.T) {
		reserved := []database.BookingTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}
		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 15

		bookingDay := ct(year, month, day, "00:00", tz)

		businessHours := []database.TimeSlot{
			{StartTime: "09:00:00", EndTime: "12:00:00"},
		}

		// With buffer=15min, the blocked period is 09:45–10:45.
		// So "09:00" and "09:15" are fine, next available is "10:45".
		expectedMorning := []string{
			"09:00",
			"09:15",
			"10:45",
			"11:00",
			"11:15",
			"11:30",
		}

		blocked := []database.BlockedTimes{}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		result := handlers.CalculateAvailableTimes(
			reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz,
		)

		assert.ElementsMatch(t, expectedMorning, result.Morning, "Morning times do not match")
		assert.Empty(t, result.Afternoon, "Afternoon should be empty")
	})

	t.Run("All day blocked time", func(t *testing.T) {
		tz, _ := time.LoadLocation("Europe/Budapest")

		bookingDay := ct(year, month, day, "00:00", tz)

		reserved := []database.BookingTime{}

		blocked := []database.BlockedTimes{
			{
				AllDay:   true,
				FromDate: ct(year, month, day, "00:00", tz),
				ToDate:   ct(year, month, day+1, "00:00", tz),
			},
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}

		businessHours := []database.TimeSlot{
			{StartTime: "09:00:00", EndTime: "17:00:00"},
		}

		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 0

		currentTime := ct(year, time.June, 15, "00:00", tz)

		result := handlers.CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.Empty(t, result.Morning, "Expected no morning slots due to full block")
		assert.Empty(t, result.Afternoon, "Expected no afternoon slots due to full block")
	})

	t.Run("Business hours blocked partially", func(t *testing.T) {
		tz, _ := time.LoadLocation("Europe/Budapest")

		bookingDay := ct(year, month, day, "00:00", tz)

		reserved := []database.BookingTime{}

		blocked := []database.BlockedTimes{
			{
				AllDay:   false,
				FromDate: ct(year, month, day-1, "10:00", tz),
				ToDate:   ct(year, month, day, "11:00", tz),
			},
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}

		businessHours := []database.TimeSlot{
			{StartTime: "08:00:00", EndTime: "12:00:00"},
		}

		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 0

		currentTime := ct(year, month, day, "00:00", tz)

		expected := []string{
			"11:00", "11:15", "11:30",
		}

		result := handlers.CalculateAvailableTimes(reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, expected, result.Morning, "Unexpected available slots for partial block")
	})

	t.Run("Blocked time overlaps only wait inside multi-phase service", func(t *testing.T) {
		tz, _ := time.LoadLocation("Europe/Budapest")

		bookingDay := ct(year, month, day, "00:00", tz)

		reserved := []database.BookingTime{}

		blocked := []database.BlockedTimes{
			{
				AllDay:   false,
				FromDate: ct(year, month, day, "09:30", tz),
				ToDate:   ct(year, month, day, "10:00", tz),
			},
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30}, // 1st active
			{PhaseType: types.ServicePhaseTypeWait, Duration: 30},   // wait
			{PhaseType: types.ServicePhaseTypeActive, Duration: 15}, // 2nd active
		}

		businessHours := []database.TimeSlot{
			{StartTime: "09:00:00", EndTime: "12:00:00"},
		}

		serviceDuration := 75
		bookingWindowMin, bufferTime := 0, 0

		currentTime := ct(year, month, day, "00:00", tz)

		expected := []string{
			"09:00", // allowed: active phases do NOT intersect block
			"10:00",
			"10:15",
			"10:30",
			"10:45",
		}

		result := handlers.CalculateAvailableTimes(
			reserved, blocked, servicePhases, serviceDuration, bufferTime, bookingWindowMin,
			bookingDay, businessHours, currentTime, tz,
		)

		assert.ElementsMatch(t, expected, result.Morning, "Unexpected available slots for WAIT-overlap test")
	})
}

func TestCalculateAvailableTimesPeriod(t *testing.T) {
	tz, _ := time.LoadLocation("Europe/Budapest")

	t.Run("Empty period", func(t *testing.T) {
		startDate := ct(2025, time.July, 1, "00:00", tz)
		endDate := ct(2025, time.June, 30, "23:59", tz)

		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 0

		blocked := []database.BlockedTimes{}

		reserved := []database.BookingTime{}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}

		businessHours := map[int][]database.TimeSlot{
			3: {
				{StartTime: "09:00:00", EndTime: "11:00:00"},
			},
		}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		results := handlers.CalculateAvailableTimesPeriod(
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

		reserved := []database.BookingTime{
			ctReserved(2025, time.July, 2, "10:00", "10:30", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: types.ServicePhaseTypeActive, Duration: 30},
		}
		serviceDuration := 30
		bookingWindowMin, bufferTime := 0, 0

		businessHours := map[int][]database.TimeSlot{
			2: {}, // Tuesday (July 1, 2025)
			3: { // Wednesday (July 2, 2025)
				{StartTime: "09:00:00", EndTime: "11:00:00"},
			},
			4: { // Thursday (July 3, 2025)
				{StartTime: "09:00:00", EndTime: "11:00:00"},
			},
		}

		blocked := []database.BlockedTimes{}

		currentTime := ct(2025, time.June, 12, "00:00", tz)

		results := handlers.CalculateAvailableTimesPeriod(
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
