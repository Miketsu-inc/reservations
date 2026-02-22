package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/teambition/rrule-go"
)

func (s *Service) generateRecurringBookings(ctx context.Context, series domain.CompleteBookingSeries, serivePhases []domain.PublicServicePhase) (int, error) {
	tz, err := time.LoadLocation(series.Timezone)
	if err != nil {
		return 0, fmt.Errorf("error parsing location from booking series: %s", err.Error())
	}

	now := time.Now().UTC()
	end := now.AddDate(0, 3, 0)

	rrule, err := rrule.StrToRRule(series.Rrule)
	if err != nil {
		return 0, fmt.Errorf("error parsing rrule string: %s", err.Error())
	}

	occurrences := rrule.Between(now, end, true)

	existingOccurrences, err := s.bookingRepo.GetExistingOccurrenceDates(ctx, series.Id, now, end)
	if err != nil {
		return 0, fmt.Errorf("could not get existing occurrence dates: %s", err.Error())
	}

	existingMap := make(map[string]bool)
	for _, date := range existingOccurrences {
		existingMap[date.Format("2006-01-02")] = true
	}

	var duration time.Duration
	for _, phase := range serivePhases {
		duration += time.Duration(phase.Duration)
	}

	duration = duration * time.Minute

	var fromDates []time.Time
	var toDates []time.Time
	for _, date := range occurrences {
		if existingMap[date.Format("2006-01-02")] {
			continue
		}

		fromDate := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, tz)
		toDate := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, tz)
		toDate = toDate.Add(duration)

		fromDates = append(fromDates, fromDate.UTC())
		toDates = append(toDates, toDate.UTC())
	}

	return s.bookingRepo.BatchCreateRecurringBookings(ctx, domain.NewRecurringBookings{
		BookingSeriesId: series.Id,
		BookingStatus:   types.BookingStatusBooked,
		BookingType:     series.BookingType,
		MerchantId:      series.MerchantId,
		EmployeeId:      series.EmployeeId,
		ServiceId:       series.ServiceId,
		LocationId:      series.LocationId,
		FromDates:       fromDates,
		ToDates:         toDates,
		Phases:          serivePhases,
		Details:         series.Details,
		Participants:    series.Participants,
	})
}
