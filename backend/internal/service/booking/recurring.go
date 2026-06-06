package booking

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/teambition/rrule-go"
)

func (s *Service) GenerateRecurringBookings(ctx context.Context, series domain.BookingSeries, seriesParticipants []domain.BookingSeriesParticipant,
	service domain.Service, generateFrom time.Time) error {
	tz, err := time.LoadLocation(series.Timezone)
	if err != nil {
		return fmt.Errorf("error parsing location from booking series: %s", err.Error())
	}

	generateUntil := generateFrom.In(tz).AddDate(0, 3, 0)

	rrule, err := rrule.StrToRRule(series.Rrule)
	if err != nil {
		return fmt.Errorf("error parsing rrule string: %s", err.Error())
	}

	occurrences := rrule.Between(generateFrom, generateUntil, false)

	if len(occurrences) == 0 {
		slog.DebugContext(ctx, fmt.Sprintf("there are no occurrences between start (%s) and end (%s) date", generateFrom, generateUntil))
		return nil
	}

	totalDuration := service.GetTotalDuration()

	bookings := make([]domain.Booking, 0, len(occurrences))
	for _, date := range occurrences {
		fromDate := date.UTC()
		toDate := fromDate.Add(totalDuration)

		bookings = append(bookings, domain.Booking{
			Status:              types.BookingStatusConfirmed,
			BookingType:         series.BookingType,
			IsRecurring:         true,
			MerchantId:          series.MerchantId,
			EmployeeId:          series.EmployeeId,
			ServiceId:           series.ServiceId,
			LocationId:          series.LocationId,
			BookingSeriesId:     &series.Id,
			SeriesOriginalDate:  &fromDate,
			FromDate:            fromDate,
			ToDate:              toDate,
			PricePerPerson:      series.PricePerPerson,
			TotalPrice:          series.TotalPrice,
			MerchantNote:        nil,
			MinParticipants:     series.MinParticipants,
			MaxParticipants:     series.MaxParticipants,
			CurrentParticipants: series.CurrentParticipants,
		})
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		bookingIds, err := s.bookingRepo.WithTx(tx).NewBookings(ctx, bookings)
		if err != nil {
			return err
		}

		bookingPhases := make([]domain.BookingPhase, 0, len(bookingIds)*len(service.Phases))
		participants := make([]domain.BookingParticipant, 0, len(bookingIds)*len(seriesParticipants))

		for i, id := range bookingIds {
			phases := service.CalculateNewBookingPhases(id, bookings[i].FromDate)
			bookingPhases = append(bookingPhases, phases...)

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
			return err
		}

		err = s.bookingRepo.WithTx(tx).NewBookingParticipants(ctx, participants)
		if err != nil {
			return err
		}

		err = s.bookingRepo.WithTx(tx).UpdateBookingSeriesGeneratedUntil(ctx, series.Id, occurrences[len(occurrences)-1])
		if err != nil {
			return err
		}

		// does not have future occurences
		if rrule.After(generateUntil, false).IsZero() {
			err = s.bookingRepo.WithTx(tx).DeactivateBookingSeries(ctx, series.Id)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

type occurrenceTimestampUpdateContext struct {
	rrule                    *rrule.RRule
	merchantTz               *time.Location
	duration                 time.Duration
	servicePhaseCount        int
	seriesOriginalDateOffset time.Duration
}

type occurrenceTimestampUpdate struct {
	BookingIds    []int
	FromDates     []time.Time
	ToDates       []time.Time
	BookingPhases []domain.BookingPhase
}

func buildOccurrenceTimestampUpdate(context occurrenceTimestampUpdateContext, futureBookings []domain.Booking, service domain.Service) (occurrenceTimestampUpdate, error) {
	futureBookingsCount := len(futureBookings)

	// fromDateOffset is necessary here to get the correct amount of occurrences
	// because the rrule should have already changed with the offset
	occurrencesStart := futureBookings[0].SeriesOriginalDate.In(context.merchantTz).Add(context.seriesOriginalDateOffset)
	occurrencesEnd := futureBookings[futureBookingsCount-1].SeriesOriginalDate.In(context.merchantTz).Add(context.seriesOriginalDateOffset)

	occurrences := context.rrule.Between(occurrencesStart, occurrencesEnd, true)

	if len(occurrences) != futureBookingsCount {
		return occurrenceTimestampUpdate{}, fmt.Errorf("number of generated occurrences (%d) and future bookings (%d) should match", len(occurrences), futureBookingsCount)
	}

	if occurrences[0].UTC().Sub(*futureBookings[0].SeriesOriginalDate) != context.seriesOriginalDateOffset {
		return occurrenceTimestampUpdate{}, fmt.Errorf("from date offset should match (%s) != (%s)", occurrences[0].UTC().Sub(*futureBookings[0].SeriesOriginalDate), context.seriesOriginalDateOffset)
	}

	var bookingIds []int
	var fromDates []time.Time
	var toDates []time.Time
	var bookingPhases []domain.BookingPhase

	for i, b := range futureBookings {
		if b.IsModifiable() {
			bookingStart := occurrences[i].UTC()

			bookingIds = append(bookingIds, b.Id)
			fromDates = append(fromDates, bookingStart)
			toDates = append(toDates, bookingStart.Add(context.duration))

			phases := service.CalculateNewBookingPhases(b.Id, bookingStart)
			bookingPhases = append(bookingPhases, phases...)
		}
	}

	return occurrenceTimestampUpdate{
		BookingIds:    bookingIds,
		FromDates:     fromDates,
		ToDates:       toDates,
		BookingPhases: bookingPhases,
	}, nil
}

func makeExistingParticipantsMap(bookingIds []int, customerIdsByBooking map[int][]uuid.UUID) map[int]map[uuid.UUID]struct{} {
	existingMap := make(map[int]map[uuid.UUID]struct{}, len(bookingIds))

	for _, id := range bookingIds {
		customerIds, ok := customerIdsByBooking[id]
		if !ok {
			existingMap[id] = make(map[uuid.UUID]struct{})
		}

		participants := make(map[uuid.UUID]struct{}, len(customerIds))
		for _, cid := range customerIds {
			participants[cid] = struct{}{}
		}

		existingMap[id] = participants
	}

	return existingMap
}

type capacityUpdate struct {
	BookingIdsExceeded []int
	BookingIdsToUpdate []int
	DeltaToInsert      []int
	ByBooking          map[int]int
}

func calculateCapacity(futureBookingsMap map[int]domain.Booking, existingParticipantsByBooking map[int]map[uuid.UUID]struct{},
	toDeleteMap, toInsertMap map[uuid.UUID]struct{}) capacityUpdate {
	var capacity capacityUpdate

	capacityDelta := make(map[int]int)

	// calculate the change required
	for bookignId, customerIds := range existingParticipantsByBooking {
		capacityDelta[bookignId] = 0

		for cid := range customerIds {
			if _, inDelete := toDeleteMap[cid]; inDelete {
				capacityDelta[bookignId] -= 1
			}
		}

		for cid := range toInsertMap {
			if _, ok := customerIds[cid]; !ok {
				capacityDelta[bookignId] += 1
			}
		}
	}

	capacity.ByBooking = make(map[int]int)
	capacity.BookingIdsExceeded = []int{}
	capacity.BookingIdsToUpdate = []int{}
	capacity.DeltaToInsert = []int{}

	// sort to make the order of booking updates deterministic
	sortedBookingIds := slices.Sorted(maps.Keys(futureBookingsMap))

	// check if the change causes current participants to exceed max
	for _, bookingId := range sortedBookingIds {
		booking := futureBookingsMap[bookingId]
		delta := capacityDelta[booking.Id]

		currentParticipants := booking.CurrentParticipants
		existingParticipantCount := len(existingParticipantsByBooking[booking.Id])

		// ideally this never happens but if it for some reason does, we need to correct it
		if currentParticipants != existingParticipantCount {
			currentParticipants = existingParticipantCount
		}

		if currentParticipants+delta > booking.MaxParticipants {
			capacity.BookingIdsExceeded = append(capacity.BookingIdsExceeded, booking.Id)
			capacity.ByBooking[booking.Id] = currentParticipants
		} else {
			if delta != 0 {
				capacity.BookingIdsToUpdate = append(capacity.BookingIdsToUpdate, booking.Id)
				capacity.DeltaToInsert = append(capacity.DeltaToInsert, delta)
			}

			capacity.ByBooking[booking.Id] = currentParticipants + delta
		}
	}

	return capacity
}

// check if all bookings were updated, update could fail due to a condition in the where clause
func checkCapacityUpdateSuccess(toUpdate []int, updated []int) []int {
	var failedToUpdate []int

	updatedCount := len(updated)

	if updatedCount != len(toUpdate) {
		updatedMap := make(map[int]struct{}, updatedCount)
		for _, id := range updated {
			updatedMap[id] = struct{}{}
		}

		for _, id := range toUpdate {
			if _, ok := updatedMap[id]; !ok {
				failedToUpdate = append(failedToUpdate, id)
			}
		}
	}

	return failedToUpdate
}

func buildOccurrenceParticipantsToInsert(bookingIds []int, requestedToInsert []uuid.UUID, existingByBooking map[int]map[uuid.UUID]struct{}) []domain.BookingParticipant {
	var toInsert []domain.BookingParticipant

	for _, bid := range bookingIds {
		for _, cId := range requestedToInsert {
			if _, existing := existingByBooking[bid][cId]; !existing {
				toInsert = append(toInsert, domain.BookingParticipant{
					BookingId:  bid,
					CustomerId: &cId,
					Status:     types.BookingStatusConfirmed,
				})
			}
		}
	}

	return toInsert
}

func calculateTotalPrices(pricePerPerson currencyx.Price, bookingIds []int, capacityByBooking map[int]int) ([]currencyx.Price, error) {
	var totalPrices []currencyx.Price

	updatedBookingIdsMap := make(map[int]struct{}, len(bookingIds))
	for _, id := range bookingIds {
		updatedBookingIdsMap[id] = struct{}{}
	}

	for _, id := range bookingIds {
		capacity := capacityByBooking[id]

		if _, inUpdated := updatedBookingIdsMap[id]; inUpdated {
			totalPrice, err := pricePerPerson.Mul(strconv.Itoa(capacity))
			if err != nil {
				return []currencyx.Price{}, err
			}

			totalPrices = append(totalPrices, currencyx.Price{Amount: totalPrice})
		}
	}

	return totalPrices, nil
}

// This whole function assumes that the series was updated before it ran and is up to date
func (s *Service) UpdateFutureBookingOccurrences(ctx context.Context, series domain.BookingSeries, service domain.Service, seriesOriginalDateOffset time.Duration,
	priceChanged bool, statusChangedToCancelled bool, requestedParticipantsToInsert []uuid.UUID, requestedParticipantsToDelete []uuid.UUID) error {

	// TODO: somehow present this to the user...
	// maybe with notifications once we implement that
	// currently this is just ignored
	var bookingsExceedingMaxParticipants []int

	// avoid repeated operations as much as possible
	// ------
	var timestampUpdateContext occurrenceTimestampUpdateContext

	timestampChanged := seriesOriginalDateOffset != time.Duration(0)

	if timestampChanged {
		parsedRrule, err := rrule.StrToRRule(series.Rrule)
		if err != nil {
			return fmt.Errorf("failed to parse existing rrule: %w", err)
		}

		merchantTz, err := time.LoadLocation(series.Timezone)
		if err != nil {
			return fmt.Errorf("failed to parse series timezone: %w", err)
		}

		timestampUpdateContext = occurrenceTimestampUpdateContext{
			rrule:                    parsedRrule,
			merchantTz:               merchantTz,
			duration:                 service.GetTotalDuration(),
			servicePhaseCount:        len(service.Phases),
			seriesOriginalDateOffset: seriesOriginalDateOffset,
		}
	}

	var toInsertMap map[uuid.UUID]struct{}
	var toDeleteMap map[uuid.UUID]struct{}
	requestedToInsertCount := len(requestedParticipantsToInsert)
	requestedToDeleteCount := len(requestedParticipantsToDelete)

	participantsChanged := requestedToInsertCount > 0 || requestedToDeleteCount > 0

	if participantsChanged {
		toInsertMap = make(map[uuid.UUID]struct{}, requestedToInsertCount)
		for _, id := range requestedParticipantsToInsert {
			toInsertMap[id] = struct{}{}
		}

		toDeleteMap = make(map[uuid.UUID]struct{}, requestedToDeleteCount)
		for _, id := range requestedParticipantsToDelete {
			toDeleteMap[id] = struct{}{}
		}
	}
	// ------

	lastBookingStart := series.Dstart.UTC()
	finshed := false

	for !finshed {
		err := s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
			// the given fromDate is not included in the future bookings
			futureBookings, err := s.bookingRepo.WithTx(tx).GetFutureSeriesBookingsWithLock(ctx, series.Id, lastBookingStart, 50)
			if err != nil {
				return fmt.Errorf("failed to fetch future series bookings: %w", err)
			}

			if len(futureBookings) == 0 {
				finshed = true
				return nil
			}

			// TODO: this is fragile as we only lock 50 at a time and one in the future could change before locking it
			// should be replaced with occurrence index once introduced
			lastBookingStart = futureBookings[len(futureBookings)-1].FromDate

			if statusChangedToCancelled {
				var bookingsToCancel []int

				for _, b := range futureBookings {
					if err = b.CanCancel(); err == nil {
						bookingsToCancel = append(bookingsToCancel, b.Id)
					}
				}

				err = s.bookingRepo.WithTx(tx).CancelBookingByMerchantBatch(ctx, bookingsToCancel)
				if err != nil {
					return err
				}

				return nil
			}

			if timestampChanged {
				timestampUpdate, err := buildOccurrenceTimestampUpdate(timestampUpdateContext, futureBookings, service)
				if err != nil {
					return err
				}

				// time stamp changed and so should the last booking start so we do not accidentally
				// query it back in the for loop's next iteration
				lastBookingStart = timestampUpdate.FromDates[len(timestampUpdate.FromDates)-1]

				err = s.bookingRepo.WithTx(tx).UpdateBookingOccurrencesBatch(ctx, timestampUpdate.BookingIds, timestampUpdate.FromDates, timestampUpdate.ToDates, series.Id)
				if err != nil {
					return fmt.Errorf("error updating booking occurrences: %w", err)
				}

				err = s.bookingRepo.WithTx(tx).DeleteBookingPhasesBatch(ctx, timestampUpdate.BookingIds)
				if err != nil {
					return fmt.Errorf("failed to delete booking phases: %w", err)
				}

				err = s.bookingRepo.WithTx(tx).NewBookingPhases(ctx, timestampUpdate.BookingPhases)
				if err != nil {
					return fmt.Errorf("failed to insert booking phases: %w", err)
				}
			}

			if participantsChanged || priceChanged {
				var futureBookingIds []int
				futureBookingsMap := make(map[int]domain.Booking)

				for _, b := range futureBookings {
					if b.IsModifiable() {
						futureBookingIds = append(futureBookingIds, b.Id)
						futureBookingsMap[b.Id] = b
					}
				}

				customerIdsByBooking, err := s.bookingRepo.WithTx(tx).GetParticipantCustomerIdsForBookings(ctx, futureBookingIds)
				if err != nil {
					return fmt.Errorf("error getting participant customer ids for bookings: %w", err)
				}

				existingParticipantsByBooking := makeExistingParticipantsMap(futureBookingIds, customerIdsByBooking)

				capacity := calculateCapacity(futureBookingsMap, existingParticipantsByBooking, toDeleteMap, toInsertMap)

				bookingsExceedingMaxParticipants = append(bookingsExceedingMaxParticipants, capacity.BookingIdsExceeded...)

				// updating the total prices relies on this as we only want to update
				// bookings which participant count updated if the participants changed
				updatedBookingIds := capacity.BookingIdsToUpdate

				if participantsChanged {
					if len(capacity.BookingIdsToUpdate) > 0 {
						updatedBookingIds, err = s.bookingRepo.WithTx(tx).UpdateParticipantCountBatch(ctx, capacity.BookingIdsToUpdate, capacity.DeltaToInsert)
						if err != nil {
							return fmt.Errorf("error updating participant count: %w", err)
						}

						failedToUpdate := checkCapacityUpdateSuccess(capacity.BookingIdsToUpdate, updatedBookingIds)
						bookingsExceedingMaxParticipants = append(bookingsExceedingMaxParticipants, failedToUpdate...)
					}

					if requestedToDeleteCount > 0 {
						err := s.bookingRepo.WithTx(tx).DeleteBookingParticipantsBatch(ctx, futureBookingIds, requestedParticipantsToDelete)
						if err != nil {
							return fmt.Errorf("failed to remove participants for future bookings: %w", err)
						}
					}

					if requestedToInsertCount > 0 {
						participantsToInsert := buildOccurrenceParticipantsToInsert(updatedBookingIds, requestedParticipantsToInsert, existingParticipantsByBooking)

						if len(participantsToInsert) > 0 {
							// we do not want to override participant statuses on conflict
							err = s.bookingRepo.WithTx(tx).UpdateBookingParticipants(ctx, participantsToInsert, false)
							if err != nil {
								return fmt.Errorf("failed to add participants: %w", err)
							}
						}
					}
				}

				if priceChanged {
					err = s.bookingRepo.WithTx(tx).UpdateBookingPricePerPersonBatch(ctx, futureBookingIds, series.PricePerPerson)
					if err != nil {
						return fmt.Errorf("failed to batch update future booking prices per person: %w", err)
					}
				}

				// If the participants changed and and the participant count update failed for a booking
				// then we should not calculate and override the totalPrice, hence the use of 'updatedBookingIds'
				if len(updatedBookingIds) > 0 {
					totalPrices, err := calculateTotalPrices(series.PricePerPerson, updatedBookingIds, capacity.ByBooking)
					if err != nil {
						return err
					}

					err = s.bookingRepo.WithTx(tx).UpdateBookingTotalPriceBatch(ctx, updatedBookingIds, totalPrices)
					if err != nil {
						return fmt.Errorf("failed to btach update future booking total prices: %w", err)
					}
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update series occurences: %w", err)
		}
	}

	return nil
}
