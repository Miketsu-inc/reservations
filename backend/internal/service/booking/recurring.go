package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/teambition/rrule-go"
)

type CompleteBookingSeries struct {
	domain.BookingSeries
	Details      domain.BookingSeriesDetails
	Participants []domain.BookingSeriesParticipant
}

func (s *Service) generateRecurringBookings(ctx context.Context, series CompleteBookingSeries, serivePhases []domain.PublicServicePhase) (int, error) {
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

	var totalDuration time.Duration
	for _, phase := range serivePhases {
		totalDuration += time.Duration(phase.Duration)
	}

	totalDuration = totalDuration * time.Minute

	var bookings []domain.Booking

	for _, date := range occurrences {
		if existingMap[date.Format("2006-01-02")] {
			continue
		}

		fromDate := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, tz)
		toDate := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, tz)

		fromDate = fromDate.UTC()
		toDate = toDate.Add(totalDuration).UTC()

		bookings = append(bookings, domain.Booking{
			Status:             types.BookingStatusBooked,
			BookingType:        series.BookingType,
			IsRecurring:        true,
			MerchantId:         series.MerchantId,
			EmployeeId:         &series.EmployeeId,
			ServiceId:          series.ServiceId,
			LocationId:         series.LocationId,
			BookingSeriesId:    &series.Id,
			SeriesOriginalDate: &fromDate,
			FromDate:           fromDate,
			ToDate:             toDate,
		})
	}

	var bookingIds []int

	err = s.txManager.WithTransaction(ctx, func(tx db.DBTX) error {
		bookingIds, err = s.bookingRepo.WithTx(tx).NewBookings(ctx, bookings)
		if err != nil {
			return err
		}

		bookingPhases := make([]domain.BookingPhase, 0, len(bookingIds)*len(serivePhases))
		bookingDetails := make([]domain.BookingDetails, len(bookingIds))
		participants := make([]domain.BookingParticipant, 0, len(bookingIds)*len(series.Participants))

		for i, id := range bookingIds {
			bookingStart := bookings[i].FromDate

			for _, phase := range serivePhases {
				phaseDuration := time.Duration(phase.Duration) * time.Minute
				bookingEnd := bookingStart.Add(phaseDuration)

				bookingPhases = append(bookingPhases, domain.BookingPhase{
					BookingId:      id,
					ServicePhaseId: phase.Id,
					FromDate:       bookingStart,
					ToDate:         bookingEnd,
				})

				bookingStart = bookingEnd
			}

			bookingDetails[i] = domain.BookingDetails{
				BookingId:           id,
				PricePerPerson:      series.Details.PricePerPerson,
				CostPerPerson:       series.Details.CostPerPerson,
				TotalPrice:          series.Details.TotalPrice,
				TotalCost:           series.Details.TotalCost,
				MerchantNote:        nil,
				MinParticipants:     series.Details.MinParticipants,
				MaxParticipants:     series.Details.MaxParticipants,
				CurrentParticipants: series.Details.CurrentParticipants,
			}

			for _, p := range series.Participants {
				participants = append(participants, domain.BookingParticipant{
					Status:       types.BookingStatusBooked,
					BookingId:    id,
					CustomerId:   p.CustomerId,
					CustomerNote: nil,
				})
			}
		}

		err = s.bookingRepo.WithTx(tx).NewBookingPhases(ctx, bookingPhases)
		if err != nil {
			return err
		}

		err = s.bookingRepo.WithTx(tx).NewBookingDetailsBatch(ctx, bookingDetails)
		if err != nil {
			return err
		}

		err = s.bookingRepo.WithTx(tx).NewBookingParticipants(ctx, participants)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return bookingIds[0], nil
}
