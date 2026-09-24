package merchant

import (
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

func hasAllDayBlock(blockedTimes []domain.BlockedTimes) bool {
	for _, b := range blockedTimes {
		if b.IsAllDay {
			return true
		}
	}

	return false
}

func filterBlockedTimesForDay(blockedTimes []domain.BlockedTimes, day time.Time, tz *time.Location) []domain.BlockedTimes {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, tz)
	dayEnd := dayStart.AddDate(0, 0, 1)
	dayYear, dayMonth, dayOfMonth := dayStart.Date()

	filtered := []domain.BlockedTimes{}
	for _, blocked := range blockedTimes {
		if blocked.IsAllDay {
			if blocked.BlockedDay == nil {
				continue
			}

			blockedYear, blockedMonth, blockedDay := blocked.BlockedDay.Date()
			if blockedYear == dayYear && blockedMonth == dayMonth && blockedDay == dayOfMonth {
				filtered = append(filtered, blocked)
			}

			continue
		}

		if blocked.FromDate == nil || blocked.ToDate == nil {
			continue
		}

		blockedFrom := blocked.FromDate.In(tz)
		blockedTo := blocked.ToDate.In(tz)
		if blockedFrom.Before(dayEnd) && blockedTo.After(dayStart) {
			filtered = append(filtered, blocked)
		}
	}

	return filtered
}

func hasNoPhaseConflict(bookingStart time.Time, servicePhases []domain.ServicePhase, blockedTimes []domain.BlockedTimes, reserved []domain.BookingSlot, bufferDuration time.Duration, merchantTz *time.Location) bool {
	phaseStart := bookingStart
	for _, phase := range servicePhases {
		phaseEnd := phaseStart.Add(phase.GetDuration())

		if phase.PhaseType == types.ServicePhaseTypeActive {
			for _, blocked := range blockedTimes {
				if !blocked.IsAllDay && blocked.FromDate != nil && blocked.ToDate != nil {
					blockedFrom := blocked.FromDate.In(merchantTz)
					blockedUntil := blocked.ToDate.In(merchantTz)

					if phaseStart.Before(blockedUntil) && phaseEnd.After(blockedFrom) {
						return false
					}
				}
			}

			for _, booking := range reserved {
				reservedFromDate := booking.FromDate.In(merchantTz).Add(-bufferDuration)
				reservedToDate := booking.ToDate.In(merchantTz).Add(bufferDuration)

				if phaseStart.Before(reservedToDate) && phaseEnd.After(reservedFromDate) {
					return false
				}
			}
		}
		phaseStart = phaseEnd
	}
	return true
}

func iterateAvailableSlots(blocked []domain.BlockedTimes, reserved []domain.BookingSlot, businessHours []domain.TimeSlot, bookingDay time.Time, servicePhases []domain.ServicePhase,
	serviceDuration, bookingWindowMin, bufferTime int, currentTime time.Time, merchantTz *time.Location, onAvailableSlot func(bookingStart time.Time) bool) bool {
	if hasAllDayBlock(blocked) {
		return true
	}

	totalDuration := time.Duration(serviceDuration) * time.Minute
	bufferDuration := time.Duration(bufferTime) * time.Minute
	bookingDeadlineDuration := time.Duration(bookingWindowMin) * time.Minute

	year, month, day := bookingDay.Date()
	stepSize := 15 * time.Minute
	now := currentTime.In(merchantTz)

	for _, slot := range businessHours {
		// buisness hours are NOT an absolute point in time,
		// their timezone should be in the same timzone as the merchant is in
		// for golang before/after to work correctly
		businessStart := time.Date(year, month, day, slot.StartTime.Hour(), slot.StartTime.Minute(), 0, 0, merchantTz)
		businessEnd := time.Date(year, month, day, slot.EndTime.Hour(), slot.EndTime.Minute(), 0, 0, merchantTz)

		bookingStart := businessStart

		for !bookingStart.Add(totalDuration).After(businessEnd) {
			if bookingStart.Before(now.Add(bookingDeadlineDuration)) {
				bookingStart = bookingStart.Add(stepSize)
				continue
			}

			if hasNoPhaseConflict(bookingStart, servicePhases, blocked, reserved, bufferDuration, merchantTz) {
				if !onAvailableSlot(bookingStart) {
					return false
				}
			}

			bookingStart = bookingStart.Add(stepSize)
		}
	}

	return true
}

type FormattedAvailableTimes struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
}

func CalculateAvailableTimes(reserved []domain.BookingSlot, blockedTimes []domain.BlockedTimes, servicePhases []domain.ServicePhase, serviceDuration int, bufferTime int,
	bookingWindowMin int, bookingDay time.Time, businessHours []domain.TimeSlot, currentTime time.Time, merchantTz *time.Location) []time.Time {

	availableTimes := []time.Time{}

	iterateAvailableSlots(blockedTimes, reserved, businessHours, bookingDay, servicePhases, serviceDuration, bookingWindowMin, bufferTime, currentTime, merchantTz,
		func(bookingStart time.Time) bool {
			availableTimes = append(availableTimes, bookingStart)
			return true
		})

	return availableTimes
}

type DayAvailability struct {
	Date        string `json:"date"`
	IsAvailable bool   `json:"is_available"`
}

func CalculateAvailableDays(reservedForPeriod []domain.BookingSlot, blockedTimes []domain.BlockedTimes, servicePhases []domain.ServicePhase, serviceDuration int, bufferTime int,
	bookingWindowMin int, startDate time.Time, endDate time.Time, businessHours domain.BusinessHours, currentTime time.Time, merchantTz *time.Location) []DayAvailability {

	results := []DayAvailability{}

	reservationsByDate := make(map[string][]domain.BookingSlot)
	for _, booking := range reservedForPeriod {
		date := booking.FromDate.In(merchantTz).Format("2006-01-02")
		reservationsByDate[date] = append(reservationsByDate[date], booking)
	}

	for d := startDate.In(merchantTz); !d.After(endDate.In(merchantTz)); d = d.AddDate(0, 0, 1) {
		businessHoursForDay := businessHours[int(d.Weekday())]

		day := d.Format("2006-01-02")

		if len(businessHoursForDay) == 0 {
			results = append(results, DayAvailability{
				Date:        day,
				IsAvailable: false,
			})
			continue
		}

		reservedForDay := reservationsByDate[day]
		blockedForDay := filterBlockedTimesForDay(blockedTimes, d, merchantTz)

		available := false

		iterateAvailableSlots(blockedForDay, reservedForDay, businessHoursForDay, d, servicePhases, serviceDuration, bookingWindowMin, bufferTime, currentTime, merchantTz,
			func(bookingStart time.Time) bool {
				available = true

				return false
			})

		results = append(results, DayAvailability{
			Date:        day,
			IsAvailable: available,
		})
	}

	return results
}

type MultiDayAvailableTimes struct {
	Date        string   `json:"date"`
	IsAvailable bool     `json:"is_available"`
	Morning     []string `json:"morning"`
	Afternoon   []string `json:"afternoon"`
}

func CalculateAvailableTimesPeriod(reservedForPeriod []domain.BookingSlot, blockedTimes []domain.BlockedTimes, servicePhases []domain.ServicePhase, serviceDuration int, bufferTime int,
	bookingWindowMin int, startDate time.Time, endDate time.Time, businessHours domain.BusinessHours, currentTime time.Time, merchantTz *time.Location) []MultiDayAvailableTimes {

	results := []MultiDayAvailableTimes{}

	reservationsByDate := make(map[string][]domain.BookingSlot)
	for _, booking := range reservedForPeriod {
		date := booking.FromDate.In(merchantTz).Format("2006-01-02")
		reservationsByDate[date] = append(reservationsByDate[date], booking)
	}

	for d := startDate.In(merchantTz); !d.After(endDate.In(merchantTz)); d = d.AddDate(0, 0, 1) {
		businessHoursForDay := businessHours[int(d.Weekday())]
		if len(businessHoursForDay) == 0 {
			continue
		}

		day := d.Format("2006-01-02")
		reservedForDay := reservationsByDate[day]

		blockedForDay := filterBlockedTimesForDay(blockedTimes, d, merchantTz)

		dayResult := CalculateAvailableTimes(reservedForDay, blockedForDay, servicePhases, serviceDuration, bufferTime, bookingWindowMin, d, businessHoursForDay, currentTime, merchantTz)

		morning := []string{}
		afternoon := []string{}

		for _, slot := range dayResult {
			formattedTime := fmt.Sprintf("%02d:%02d", slot.Hour(), slot.Minute())
			if slot.Hour() < 12 {
				morning = append(morning, formattedTime)
			} else {
				afternoon = append(afternoon, formattedTime)
			}
		}

		isAvailable := len(morning) > 0 || len(afternoon) > 0

		results = append(results, MultiDayAvailableTimes{
			Date:        d.Format("2006-01-02"),
			IsAvailable: isAvailable,
			Morning:     morning,
			Afternoon:   afternoon,
		})
	}

	return results
}

func IsValidBookingSlot(appointmentSlot domain.TimeSlot, reservedTimes []domain.BookingSlot, blockedTimes []domain.BlockedTimes, servicePhases []domain.ServicePhase,
	businessHours []domain.TimeSlot, bufferTime, bookingWindowMin, bookingWindowMax int, currentTime time.Time, merchantTz *time.Location) bool {

	bufferDuration := time.Duration(bufferTime) * time.Minute
	bookingDeadlineDuration := time.Duration(bookingWindowMin) * time.Minute

	now := currentTime.In(merchantTz)

	if appointmentSlot.StartTime.Before(now.Add(bookingDeadlineDuration)) || appointmentSlot.EndTime.After(now.AddDate(0, bookingWindowMax, 0)) {
		return false
	}

	if hasAllDayBlock(blockedTimes) {
		return false
	}

	year, month, day := appointmentSlot.StartTime.In(merchantTz).Date()
	isWithinBusinessHours := false

	for _, slot := range businessHours {
		businessStart := time.Date(year, month, day, slot.StartTime.Hour(), slot.StartTime.Minute(), 0, 0, merchantTz)
		businessEnd := time.Date(year, month, day, slot.EndTime.Hour(), slot.EndTime.Minute(), 0, 0, merchantTz)

		if (appointmentSlot.StartTime.After(businessStart) || appointmentSlot.StartTime.Equal(businessStart)) &&
			(appointmentSlot.EndTime.Before(businessEnd) || appointmentSlot.EndTime.Equal(businessEnd)) {
			isWithinBusinessHours = true
			break
		}

	}

	if !isWithinBusinessHours {
		return false
	}

	return hasNoPhaseConflict(appointmentSlot.StartTime, servicePhases, blockedTimes, reservedTimes, bufferDuration, merchantTz)
}
