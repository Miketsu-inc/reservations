package booking

import (
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/cmd/database"
)

type FormattedAvailableTimes struct {
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
}

func CalculateAvailableTimes(reserved []database.BookingTime, servicePhases []database.PublicServicePhase, serviceDuration int, BufferTime int,
	BookingWindowMin int, bookingDay time.Time, businessHours []database.TimeSlot, currentTime time.Time, merchantTz *time.Location) FormattedAvailableTimes {

	year, month, day := bookingDay.Date()
	totalDuration := time.Duration(serviceDuration) * time.Minute
	bufferDuration := time.Duration(BufferTime) * time.Minute
	bookingDeadlineDuration := time.Duration(BookingWindowMin) * time.Minute

	morning := []string{}
	afternoon := []string{}

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

				if phase.PhaseType == "active" {

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
	Date      string   `json:"date"`
	Morning   []string `json:"morning"`
	Afternoon []string `json:"afternoon"`
}

func CalculateAvailableTimesPeriod(reservedForPeriod []database.BookingTime, servicePhases []database.PublicServicePhase, serviceDuration int, bufferTime int, bookingindowMin int,
	startDate time.Time, endDate time.Time, businessHours map[int][]database.TimeSlot, currentTime time.Time, merchantTz *time.Location) []MultiDayAvailableTimes {

	results := []MultiDayAvailableTimes{}

	reservationsByDate := make(map[string][]database.BookingTime)
	for _, booking := range reservedForPeriod {
		date := booking.From_date.In(merchantTz).Format("2006-01-02")
		reservationsByDate[date] = append(reservationsByDate[date], booking)
	}

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		businessHoursForDay := businessHours[int(d.Weekday())]
		if len(businessHoursForDay) == 0 {
			continue
		}

		day := d.Format("2006-01-02")
		reservedForDay := reservationsByDate[day]

		dayResult := CalculateAvailableTimes(reservedForDay, servicePhases, serviceDuration, bufferTime, bookingindowMin, d, businessHoursForDay, currentTime, merchantTz)

		results = append(results, MultiDayAvailableTimes{
			Date:      d.Format("2006-01-02"),
			Morning:   dayResult.Morning,
			Afternoon: dayResult.Afternoon,
		})
	}

	return results
}
