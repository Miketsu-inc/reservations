package booking_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/miketsu-inc/reservations/backend/cmd/booking"
	"github.com/miketsu-inc/reservations/backend/cmd/database"
	"github.com/stretchr/testify/assert"
)

func ct(year int, month time.Month, day int, timeStr string, loc *time.Location) time.Time {
	t, _ := time.Parse("15:04", timeStr)
	return time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, loc)
}

func ctReserved(year int, month time.Month, day int, start, end string, loc *time.Location) database.AppointmentTime {
	return database.AppointmentTime{
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
		reserved := []database.AppointmentTime{}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: "active", Duration: 30},
		}
		serviceDuration := 30

		bookingDay := ct(2025, time.July, 1, "00:00", tz)

		businessHours := []database.TimeSlot{
			{StartTime: "09:30:00", EndTime: "11:30:00"},
			{StartTime: "13:00:00", EndTime: "16:15:00"},
		}

		expectedMorning := []string{"09:30", "09:45", "10:00", "10:15", "10:30", "10:45", "11:00"}
		expectedAfternoon := []string{"13:00", "13:15", "13:30", "13:45", "14:00", "14:15", "14:30", "14:45", "15:00", "15:15", "15:30", "15:45"}

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		result := booking.CalculateAvailableTimes(reserved, servicePhases, serviceDuration, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, expectedMorning, result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, expectedAfternoon, result.Afternoon, "Afternoon times do not match")
	})

	t.Run("One active phase", func(t *testing.T) {
		reserved := []database.AppointmentTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:00", "11:45", tz),
			ctReserved(year, month, day, "13:00", "14:00", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: "active", Duration: 60},
		}
		serviceDuration := 60

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

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		result := booking.CalculateAvailableTimes(reserved, servicePhases, serviceDuration, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, formatTimes(expectedMorning), result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, formatTimes(expectedAfternoon), result.Afternoon, "Afternoon times do not match")
	})

	t.Run("Mutliple phases with wait at the start", func(t *testing.T) {
		reserved := []database.AppointmentTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:15", "11:30", tz),
			ctReserved(year, month, day, "13:00", "15:00", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: "wait", Duration: 30},
			{PhaseType: "active", Duration: 15},
		}
		serviceDuration := 45

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

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		result := booking.CalculateAvailableTimes(reserved, servicePhases, serviceDuration, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, formatTimes(expectedMorning), result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, formatTimes(expectedAfternoon), result.Afternoon, "Afternoon times do not match")
	})

	t.Run("Mutliple phases with wait in the middle", func(t *testing.T) {
		reserved := []database.AppointmentTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:15", "11:45", tz),
			ctReserved(year, month, day, "13:00", "14:00", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: "active", Duration: 15},
			{PhaseType: "wait", Duration: 30},
			{PhaseType: "active", Duration: 45},
		}
		serviceDuration := 90

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

		currentTime := ct(2025, time.June, 12, "00:00", time.UTC)

		result := booking.CalculateAvailableTimes(reserved, servicePhases, serviceDuration, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, formatTimes(expectedMorning), result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, formatTimes(expectedAfternoon), result.Afternoon, "Afternoon times do not match")
	})

	t.Run("Close current time", func(t *testing.T) {
		reserved := []database.AppointmentTime{
			ctReserved(year, month, day, "10:00", "10:30", tz),
			ctReserved(year, month, day, "11:00", "11:45", tz),
			ctReserved(year, month, day, "13:00", "14:00", tz),
		}

		servicePhases := []database.PublicServicePhase{
			{PhaseType: "active", Duration: 15},
			{PhaseType: "wait", Duration: 30},
			{PhaseType: "active", Duration: 45},
		}
		serviceDuration := 90

		bookingDay := ct(year, month, day, "00:00", tz)

		businessHours := []database.TimeSlot{
			{StartTime: "09:00:00", EndTime: "16:00:00"},
		}

		expectedMorning := []time.Time{}
		expectedAfternoon := []time.Time{
			ct(year, month, day, "14:30", tz),
		}

		currentTime := ct(2025, time.July, 1, "14:20", tz)

		result := booking.CalculateAvailableTimes(reserved, servicePhases, serviceDuration, bookingDay, businessHours, currentTime, tz)

		assert.ElementsMatch(t, formatTimes(expectedMorning), result.Morning, "Morning times do not match")
		assert.ElementsMatch(t, formatTimes(expectedAfternoon), result.Afternoon, "Afternoon times do not match")
	})
}
