package merchant

import (
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type FormattedAvailableTimes struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
}

func CalculateAvailableTimes(reserved []domain.BookingTime, blockedTimes []domain.BlockedTimes, servicePhases []domain.PublicServicePhase, serviceDuration int, BufferTime int,
	BookingWindowMin int, bookingDay time.Time, businessHours []domain.TimeSlot, currentTime time.Time, merchantTz *time.Location) FormattedAvailableTimes {

	year, month, day := bookingDay.Date()
	totalDuration := time.Duration(serviceDuration) * time.Minute
	bufferDuration := time.Duration(BufferTime) * time.Minute
	bookingDeadlineDuration := time.Duration(BookingWindowMin) * time.Minute

	morning := []string{}
	afternoon := []string{}

	for _, blocked := range blockedTimes {
		if blocked.AllDay {
			return FormattedAvailableTimes{
				Morning:   morning,
				Afternoon: afternoon,
			}
		}
	}

	now := currentTime.In(merchantTz)

	stepSize := 15 * time.Minute

	for _, slot := range businessHours {
		startTime, _ := time.Parse("15:04:05", slot.StartTime)
		endTime, _ := time.Parse("15:04:05", slot.EndTime)

		// buisness hours are NOT an absolute point in time,
		// their timezone should be in the same timzone as the merchant is in
		// for golang before/after to work correctly
		businessStart := time.Date(year, month, day, startTime.Hour(), startTime.Minute(), 0, 0, merchantTz)
		businessEnd := time.Date(year, month, day, endTime.Hour(), endTime.Minute(), 0, 0, merchantTz)

		bookingStart := businessStart

		for bookingStart.Add(totalDuration).Before(businessEnd) || bookingStart.Add(totalDuration).Equal(businessEnd) {
			if bookingStart.Before(now.Add(bookingDeadlineDuration)) {
				bookingStart = bookingStart.Add(stepSize)
				continue
			}

			available := true

			phaseStart := bookingStart
			for _, phase := range servicePhases {
				phaseDuration := time.Duration(phase.Duration) * time.Minute
				phaseEnd := phaseStart.Add(phaseDuration)

				if phase.PhaseType == types.ServicePhaseTypeActive {

					for _, blocked := range blockedTimes {
						if !blocked.AllDay {
							blockedFrom := blocked.FromDate.In(merchantTz)
							blockedTo := blocked.ToDate.In(merchantTz)

							if phaseStart.Before(blockedTo) && phaseEnd.After(blockedFrom) {
								bookingStart = bookingStart.Add(stepSize)

								available = false
								break
							}
						}
					}

					if !available {
						break
					}

					for _, booking := range reserved {
						reservedFromDate := booking.From_date.In(merchantTz).Add(-bufferDuration)
						reservedToDate := booking.To_date.In(merchantTz).Add(bufferDuration)

						if phaseStart.Before(reservedToDate) && phaseEnd.After(reservedFromDate) {
							bookingStart = bookingStart.Add(stepSize)

							available = false
							break
						}
					}
				}

				if !available {
					break
				}

				phaseStart = phaseEnd
			}

			if available {
				formattedTime := fmt.Sprintf("%02d:%02d", bookingStart.Hour(), bookingStart.Minute())

				if bookingStart.Hour() < 12 {
					morning = append(morning, formattedTime)
				} else if bookingStart.Hour() >= 12 {
					afternoon = append(afternoon, formattedTime)
				}

				bookingStart = bookingStart.Add(stepSize)
			}
		}
	}

	availableTimes := FormattedAvailableTimes{
		Morning:   morning,
		Afternoon: afternoon,
	}

	return availableTimes
}

type MultiDayAvailableTimes struct {
	Date        string   `json:"date"`
	IsAvailable bool     `json:"is_available"`
	Morning     []string `json:"morning"`
	Afternoon   []string `json:"afternoon"`
}

func filterBlockedTimesForDay(blockedTimes []domain.BlockedTimes, day time.Time, tz *time.Location) []domain.BlockedTimes {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, tz)
	dayEnd := dayStart.AddDate(0, 0, 1)

	filtered := []domain.BlockedTimes{}
	for _, blocked := range blockedTimes {
		blockedFrom := blocked.FromDate.In(tz)
		blockedTo := blocked.ToDate.In(tz)
		if blockedFrom.Before(dayEnd) && blockedTo.After(dayStart) {
			filtered = append(filtered, blocked)
		}
	}

	return filtered
}

func CalculateAvailableTimesPeriod(reservedForPeriod []domain.BookingTime, blockedTimes []domain.BlockedTimes, servicePhases []domain.PublicServicePhase, serviceDuration int, bufferTime int, bookingindowMin int,
	startDate time.Time, endDate time.Time, businessHours map[int][]domain.TimeSlot, currentTime time.Time, merchantTz *time.Location) []MultiDayAvailableTimes {

	results := []MultiDayAvailableTimes{}

	reservationsByDate := make(map[string][]domain.BookingTime)
	for _, booking := range reservedForPeriod {
		date := booking.From_date.In(merchantTz).Format("2006-01-02")
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

		dayResult := CalculateAvailableTimes(reservedForDay, blockedForDay, servicePhases, serviceDuration, bufferTime, bookingindowMin, d, businessHoursForDay, currentTime, merchantTz)

		isAvailable := len(dayResult.Morning) > 0 || len(dayResult.Afternoon) > 0

		results = append(results, MultiDayAvailableTimes{
			Date:        d.Format("2006-01-02"),
			IsAvailable: isAvailable,
			Morning:     dayResult.Morning,
			Afternoon:   dayResult.Afternoon,
		})
	}

	return results
}
