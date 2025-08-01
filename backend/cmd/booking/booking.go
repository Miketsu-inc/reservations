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

func CalculateAvailableTimes(reserved []database.AppointmentTime, servicePhases []database.PublicServicePhase, serviceDuration int,
	bookingDay time.Time, businessHours []database.TimeSlot, currentTime time.Time, merchantTz *time.Location) FormattedAvailableTimes {

	year, month, day := bookingDay.Date()
	totalDuration := time.Duration(serviceDuration) * time.Minute

	morning := []string{}
	afternoon := []string{}

	now := currentTime.In(merchantTz)
	isToday := bookingDay.Format("2006-01-02") == currentTime.Format("2006-01-02")

	stepSize := 15 * time.Minute

	for _, slot := range businessHours {
		startTime, _ := time.Parse("15:04:05", slot.StartTime)
		endTime, _ := time.Parse("15:04:05", slot.EndTime)

		// buisness hours are NOT an absolute point in time,
		// their timezone should be in the same timzone as the merchant is in
		// for golang before/after to work correctly
		businessStart := time.Date(year, month, day, startTime.Hour(), startTime.Minute(), 0, 0, merchantTz)
		businessEnd := time.Date(year, month, day, endTime.Hour(), endTime.Minute(), 0, 0, merchantTz)

		appStart := businessStart

		for appStart.Add(totalDuration).Before(businessEnd) || appStart.Add(totalDuration).Equal(businessEnd) {
			if isToday && appStart.Before(now) {
				appStart = appStart.Add(stepSize)
				continue
			}

			available := true

			phaseStart := appStart
			for _, phase := range servicePhases {
				phaseDuration := time.Duration(phase.Duration) * time.Minute
				phaseEnd := phaseStart.Add(phaseDuration)

				if phase.PhaseType == "active" {

					for _, appt := range reserved {
						reservedFromDate := appt.From_date.In(merchantTz)
						reservedToDate := appt.To_date.In(merchantTz)

						if phaseStart.Before(reservedToDate) && phaseEnd.After(reservedFromDate) {
							appStart = appStart.Add(stepSize)

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
				formattedTime := fmt.Sprintf("%02d:%02d", appStart.Hour(), appStart.Minute())

				if appStart.Hour() < 12 {
					morning = append(morning, formattedTime)
				} else if appStart.Hour() >= 12 {
					afternoon = append(afternoon, formattedTime)
				}

				appStart = appStart.Add(stepSize)
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

func CalculateAvailableTimesPeriod(reservedForPeriod []database.AppointmentTime, servicePhases []database.PublicServicePhase, serviceDuration int,
	startDate time.Time, endDate time.Time, businessHours map[int][]database.TimeSlot, currentTime time.Time, merchantTz *time.Location) []MultiDayAvailableTimes {

	results := []MultiDayAvailableTimes{}

	reservationsByDate := make(map[string][]database.AppointmentTime)
	for _, appt := range reservedForPeriod {
		date := appt.From_date.In(merchantTz).Format("2006-01-02")
		reservationsByDate[date] = append(reservationsByDate[date], appt)
	}

	for d := startDate; d.After(endDate) == false; d = d.AddDate(0, 0, 1) {
		businessHoursForDay := businessHours[int(d.Weekday())]
		if len(businessHoursForDay) == 0 {
			continue
		}

		day := d.Format("2006-01-02")
		reservedForDay := reservationsByDate[day]

		dayResult := CalculateAvailableTimes(reservedForDay, servicePhases, serviceDuration, d, businessHoursForDay, currentTime, merchantTz)

		results = append(results, MultiDayAvailableTimes{
			Date:      d.Format("2006-01-02"),
			Morning:   dayResult.Morning,
			Afternoon: dayResult.Afternoon,
		})
	}

	return results
}
