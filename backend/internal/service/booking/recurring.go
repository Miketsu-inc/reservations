package booking

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
	"github.com/teambition/rrule-go"
)

func (s *Service) GenerateRecurringBookings(ctx context.Context, tx db.DBTX, series domain.BookingSeries, seriesDetails domain.BookingSeriesDetails,
	seriesParticipants []domain.BookingSeriesParticipant, serivePhases []domain.PublicServicePhase, generateFrom time.Time) (int, error) {
	tz, err := time.LoadLocation(series.Timezone)
	if err != nil {
		return 0, fmt.Errorf("error parsing location from booking series: %s", err.Error())
	}

	generateUntil := generateFrom.UTC().AddDate(0, 3, 0)

	rrule, err := rrule.StrToRRule(series.Rrule)
	if err != nil {
		return 0, fmt.Errorf("error parsing rrule string: %s", err.Error())
	}

	occurrences := rrule.Between(generateFrom, generateUntil, false)

	if len(occurrences) == 0 {
		slog.DebugContext(ctx, fmt.Sprintf("there are no occurrences between start (%s) and end (%s) date", generateFrom, generateUntil))
		return 0, nil
	}

	var totalDuration time.Duration
	for _, phase := range serivePhases {
		totalDuration += time.Duration(phase.Duration) * time.Minute
	}

	bookings := make([]domain.Booking, 0, len(occurrences))
	for _, date := range occurrences {
		fromDate := time.Date(date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), 0, 0, tz).UTC()
		toDate := fromDate.Add(totalDuration)

		bookings = append(bookings, domain.Booking{
			Status:             types.BookingStatusConfirmed,
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

	bookingIds, err := s.bookingRepo.WithTx(tx).NewBookings(ctx, bookings)
	if err != nil {
		return 0, err
	}

	bookingPhases := make([]domain.BookingPhase, 0, len(bookingIds)*len(serivePhases))
	bookingDetails := make([]domain.BookingDetails, len(bookingIds))
	participants := make([]domain.BookingParticipant, 0, len(bookingIds)*len(seriesParticipants))

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
			PricePerPerson:      seriesDetails.PricePerPerson,
			CostPerPerson:       seriesDetails.CostPerPerson,
			TotalPrice:          seriesDetails.TotalPrice,
			TotalCost:           seriesDetails.TotalCost,
			MerchantNote:        nil,
			MinParticipants:     seriesDetails.MinParticipants,
			MaxParticipants:     seriesDetails.MaxParticipants,
			CurrentParticipants: seriesDetails.CurrentParticipants,
		}

		for _, p := range seriesParticipants {
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
		return 0, err
	}

	err = s.bookingRepo.WithTx(tx).NewBookingDetailsBatch(ctx, bookingDetails)
	if err != nil {
		return 0, err
	}

	err = s.bookingRepo.WithTx(tx).NewBookingParticipants(ctx, participants)
	if err != nil {
		return 0, err
	}

	err = s.bookingRepo.WithTx(tx).UpdateBookingSeriesGeneratedUntil(ctx, series.Id, occurrences[len(occurrences)-1])
	if err != nil {
		return 0, err
	}

	// does not have future occurences
	if rrule.After(generateUntil, false).IsZero() {
		err = s.bookingRepo.WithTx(tx).DeactivateBookingSeries(ctx, series.Id)
		if err != nil {
			return 0, err
		}
	}

	return bookingIds[0], nil
}
