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
	"github.com/miketsu-inc/reservations/backend/internal/jobs/args"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/riverqueue/river"
	"github.com/teambition/rrule-go"
)

func (s *Service) GenerateRecurringBookings(ctx context.Context, series domain.BookingSeries, seriesParticipants []domain.BookingSeriesParticipant,
	service domain.Service, generateFrom time.Time, fromOccurrenceIndex int) error {
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
	for i, date := range occurrences {
		fromDate := date.UTC()
		toDate := fromDate.Add(totalDuration)

		occurrenceIndex := fromOccurrenceIndex + i

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
			OccurrenceIndex:     &occurrenceIndex,
			SeriesVersion:       &series.Version,
		})
	}

	return s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
		bookingIds, err := s.bookingRepo.WithTx(tx).NewBookings(ctx, bookings)
		if err != nil {
			return err
		}

		bookingPhases := make([]domain.BookingPhase, 0, len(bookingIds)*len(service.Phases))
		participants := make([]domain.BookingParticipant, 0, len(bookingIds)*len(seriesParticipants))
		var reminderInsertParams []river.InsertManyParams

		for i, id := range bookingIds {
			fromDate := bookings[i].FromDate

			phases := service.CalculateNewBookingPhases(id, fromDate)
			bookingPhases = append(bookingPhases, phases...)

			for _, p := range seriesParticipants {
				participants = append(participants, domain.BookingParticipant{
					Status:       types.BookingStatusBooked,
					BookingId:    id,
					CustomerId:   p.CustomerId,
					CustomerNote: nil,
				})

				if p.CustomerId != nil {
					reminderInsertParams = append(reminderInsertParams, river.InsertManyParams{
						Args: args.BookingReminderEmail{
							BookingId:        id,
							CustomerId:       *p.CustomerId,
							ExpectedFromDate: fromDate,
						}, InsertOpts: &river.InsertOpts{
							ScheduledAt: fromDate.Add(-24 * time.Hour),
						},
					})
				}
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

		if len(reminderInsertParams) > 0 {
			_, err = s.enqueuer.InsertManyFastTx(ctx, tx, reminderInsertParams)
			if err != nil {
				return err
			}
		}

		// does not have future occurrences
		if rrule.After(generateUntil, false).IsZero() {
			err = s.bookingRepo.WithTx(tx).DeactivateBookingSeries(ctx, series.Id)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func getNextOccurrence(rrule *rrule.RRule, prev time.Time) (time.Time, error) {
	// we do not allow secondly rrule frequency so this should be fine
	d := rrule.After(prev.Add(1*time.Second), true)
	if d.IsZero() {
		return time.Time{}, fmt.Errorf("rrule ended after: %s", prev)
	}
	return d, nil
}

type occurrenceTimestampUpdateContext struct {
	rrule                    *rrule.RRule
	merchantTz               *time.Location
	duration                 time.Duration
	servicePhaseCount        int
	seriesOriginalDateOffset time.Duration
	seriesVersion            int
}

type occurrenceTimestampUpdate struct {
	BookingIds    []int
	FromDates     []time.Time
	ToDates       []time.Time
	BookingPhases []domain.BookingPhase
}

func buildOccurrenceTimestampUpdate(context occurrenceTimestampUpdateContext, futureBookings []domain.Booking, service domain.Service, lastOccurrenceDate time.Time) (occurrenceTimestampUpdate, error) {
	var bookingIds []int
	var fromDates []time.Time
	var toDates []time.Time
	var bookingPhases []domain.BookingPhase

	nextOccurrence := lastOccurrenceDate.In(context.merchantTz)

	for _, b := range futureBookings {
		occurrence, err := getNextOccurrence(context.rrule, nextOccurrence)
		if err != nil {
			return occurrenceTimestampUpdate{}, fmt.Errorf("series ended earlier than expected: %w", err)
		}

		nextOccurrence = occurrence

		if *b.SeriesVersion >= context.seriesVersion {
			continue
		}

		if b.IsModifiable() {
			bookingStart := occurrence.UTC()

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

func buildCancellationEmailParams(seriesParticipants map[uuid.UUID]struct{}, customerIdsByBooking map[int][]uuid.UUID, reason string) []river.InsertManyParams {
	var params []river.InsertManyParams

	for bookingId, customerIds := range customerIdsByBooking {
		for _, cid := range customerIds {
			// for series participants we send cancellation emails for the entire series in one email
			// which is handled in the UpdateByMerchant function
			if _, inSeries := seriesParticipants[cid]; !inSeries {
				params = append(params, river.InsertManyParams{
					Args: args.BookingCancellationEmail{
						BookingId:          bookingId,
						CustomerId:         cid,
						CancellationReason: reason,
					},
				})
			}
		}
	}

	return params
}

func buildModificationEmailParams(seriesParticipants map[uuid.UUID]struct{}, bookings []domain.Booking, customerIdsByBooking map[int][]uuid.UUID) []river.InsertManyParams {
	var params []river.InsertManyParams

	for _, b := range bookings {
		if !b.IsModifiable() {
			continue
		}

		for _, cid := range customerIdsByBooking[b.Id] {
			// for series participants we send modification emails for the entire series in one email
			// which is handled in the UpdateByMerchant function
			if _, inSeries := seriesParticipants[cid]; !inSeries {
				params = append(params, river.InsertManyParams{
					Args: args.BookingModificationEmail{
						BookingId:  b.Id,
						CustomerId: cid,
						// TODO: replace once it's on booking
						OldServiceName: "",
						OldFromDate:    b.FromDate,
						OldToDate:      b.ToDate,
					},
				})
			}
		}
	}

	return params
}

func buildNewParticipantReminderEmailParams(participants []domain.BookingParticipant, fromDateByBooking map[int]time.Time) []river.InsertManyParams {
	var params []river.InsertManyParams

	for _, p := range participants {
		if p.CustomerId == nil {
			return nil
		}

		if fromDate, ok := fromDateByBooking[p.BookingId]; ok {
			params = append(params, river.InsertManyParams{
				Args: args.BookingReminderEmail{
					BookingId:        p.BookingId,
					CustomerId:       *p.CustomerId,
					ExpectedFromDate: fromDate,
				},
				InsertOpts: &river.InsertOpts{
					ScheduledAt: fromDate.Add(-24 * time.Hour),
				},
			})
		}
	}

	return params
}

func buildReminderEmailParams(bookingIds []int, fromDates []time.Time, customerIdsByBooking map[int][]uuid.UUID) []river.InsertManyParams {
	var params []river.InsertManyParams

	for i, bookingId := range bookingIds {
		fromDate := fromDates[i]

		for _, cid := range customerIdsByBooking[bookingId] {
			params = append(params, river.InsertManyParams{
				Args: args.BookingReminderEmail{
					BookingId:        bookingId,
					CustomerId:       cid,
					ExpectedFromDate: fromDate,
				},
				InsertOpts: &river.InsertOpts{
					ScheduledAt: fromDate.Add(-24 * time.Hour),
				},
			})
		}
	}

	return params
}

// This whole function assumes that the series was updated before it ran and is up to date
func (s *Service) UpdateFutureBookingOccurrences(ctx context.Context, series domain.BookingSeries, seriesParticipants []domain.BookingSeriesParticipant, service domain.Service,
	seriesOriginalDateOffset time.Duration, priceChanged bool, statusChangedToCancelled bool, cancellation_reason string, occurrenceIndex int, requestedParticipantsToInsert []uuid.UUID, requestedParticipantsToDelete []uuid.UUID) error {

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
			seriesVersion:            series.Version,
		}
	}

	var toInsertMap map[uuid.UUID]struct{}
	var toDeleteMap map[uuid.UUID]struct{}

	participantsChanged := len(requestedParticipantsToInsert) > 0 || len(requestedParticipantsToDelete) > 0

	if participantsChanged {
		toInsertMap = make(map[uuid.UUID]struct{}, len(requestedParticipantsToInsert))
		for _, id := range requestedParticipantsToInsert {
			toInsertMap[id] = struct{}{}
		}

		toDeleteMap = make(map[uuid.UUID]struct{}, len(requestedParticipantsToDelete))
		for _, id := range requestedParticipantsToDelete {
			toDeleteMap[id] = struct{}{}
		}
	}

	seriesParticipantsMap := make(map[uuid.UUID]struct{}, len(seriesParticipants))
	for _, p := range seriesParticipants {
		if p.CustomerId != nil {
			seriesParticipantsMap[*p.CustomerId] = struct{}{}
		}
	}
	// ------

	lastOccurrenceIndex := occurrenceIndex
	finshed := false

	for !finshed {
		err := s.txManager.WithTransaction(ctx, func(tx pgx.Tx) error {
			// the given occurrence index is not included
			futureBookings, err := s.bookingRepo.WithTx(tx).GetFutureSeriesBookingsWithLock(ctx, series.Id, lastOccurrenceIndex, 50)
			if err != nil {
				return fmt.Errorf("failed to fetch future series bookings: %w", err)
			}

			if len(futureBookings) == 0 {
				finshed = true
				return nil
			}

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

				customerIdsByBooking, err := s.bookingRepo.WithTx(tx).GetParticipantCustomerIdsForBookings(ctx, bookingsToCancel)
				if err != nil {
					return fmt.Errorf("error getting participant customer ids for bookings: %w", err)
				}

				cancellationParams := buildCancellationEmailParams(seriesParticipantsMap, customerIdsByBooking, cancellation_reason)
				if len(cancellationParams) > 0 {
					_, err := s.enqueuer.InsertManyFastTx(ctx, tx, cancellationParams)
					if err != nil {
						return fmt.Errorf("failed to schedule cancellation email: %w", err)
					}
				}

				lastOccurrenceIndex = *futureBookings[len(futureBookings)-1].OccurrenceIndex

				return nil
			}

			var futureBookingIds []int
			futureBookingsMap := make(map[int]domain.Booking)

			for _, b := range futureBookings {
				if b.IsModifiable() {
					futureBookingIds = append(futureBookingIds, b.Id)
					futureBookingsMap[b.Id] = b
				}
			}

			// needed to send booking reminders with the new from date if timestamp was changed
			fromDateByBooking := make(map[int]time.Time, len(futureBookingsMap))
			for id, b := range futureBookingsMap {
				fromDateByBooking[id] = b.FromDate
			}

			var customerIdsByBooking map[int][]uuid.UUID

			if timestampChanged || participantsChanged || priceChanged {
				customerIdsByBooking, err = s.bookingRepo.WithTx(tx).GetParticipantCustomerIdsForBookings(ctx, futureBookingIds)
				if err != nil {
					return fmt.Errorf("error getting participant customer ids for bookings: %w", err)
				}
			}

			if timestampChanged {
				lastOccurrenceDate, err := s.bookingRepo.WithTx(tx).GetSeriesOccurrenceDateByIndex(ctx, lastOccurrenceIndex)
				if err != nil {
					return fmt.Errorf("error retrieving last occurrence date: %w", err)
				}

				timestampUpdate, err := buildOccurrenceTimestampUpdate(timestampUpdateContext, futureBookings, service, lastOccurrenceDate)
				if err != nil {
					return err
				}

				if len(timestampUpdate.BookingIds) > 0 {
					err = s.bookingRepo.WithTx(tx).UpdateBookingOccurrencesBatch(ctx, timestampUpdate.BookingIds, timestampUpdate.FromDates, timestampUpdate.ToDates, series.Id, series.Version)
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

					for i, id := range timestampUpdate.BookingIds {
						fromDateByBooking[id] = timestampUpdate.FromDates[i]
					}

					reminderParams := buildReminderEmailParams(timestampUpdate.BookingIds, timestampUpdate.FromDates, customerIdsByBooking)
					if len(reminderParams) > 0 {
						_, err := s.enqueuer.InsertManyFastTx(ctx, tx, reminderParams)
						if err != nil {
							return fmt.Errorf("failed to schedule booking reminder emails: %w", err)
						}
					}
				}
			}

			lastOccurrenceIndex = *futureBookings[len(futureBookings)-1].OccurrenceIndex

			if participantsChanged || priceChanged {
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

					if len(requestedParticipantsToDelete) > 0 {
						err := s.bookingRepo.WithTx(tx).DeleteBookingParticipantsBatch(ctx, futureBookingIds, requestedParticipantsToDelete)
						if err != nil {
							return fmt.Errorf("failed to remove participants for future bookings: %w", err)
						}

						customerIdsToDeleteByBooking := make(map[int][]uuid.UUID, len(futureBookingIds))
						for _, bid := range futureBookingIds {
							customerIdsToDeleteByBooking[bid] = requestedParticipantsToDelete
						}

						cancellationParams := buildCancellationEmailParams(seriesParticipantsMap, customerIdsToDeleteByBooking, "")
						if len(cancellationParams) > 0 {
							_, err = s.enqueuer.InsertManyFastTx(ctx, tx, cancellationParams)
							if err != nil {
								return fmt.Errorf("failed to schedule cancellation emails: %w", err)
							}
						}
					}

					if len(requestedParticipantsToInsert) > 0 {
						participantsToInsert := buildOccurrenceParticipantsToInsert(updatedBookingIds, requestedParticipantsToInsert, existingParticipantsByBooking)

						if len(participantsToInsert) > 0 {
							// we do not want to override participant statuses on conflict
							err = s.bookingRepo.WithTx(tx).UpdateBookingParticipants(ctx, participantsToInsert, false)
							if err != nil {
								return fmt.Errorf("failed to add participants: %w", err)
							}

							reminderParams := buildNewParticipantReminderEmailParams(participantsToInsert, fromDateByBooking)
							if len(reminderParams) > 0 {
								_, err := s.enqueuer.InsertManyFastTx(ctx, tx, reminderParams)
								if err != nil {
									return fmt.Errorf("failed to schedule new participant reminder emails: %w", err)
								}
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

				// If the participants changed and the participant count update failed for a booking
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

				if timestampChanged || priceChanged {
					modificationParams := buildModificationEmailParams(seriesParticipantsMap, futureBookings, customerIdsByBooking)
					if len(modificationParams) > 0 {
						_, err = s.enqueuer.InsertManyFastTx(ctx, tx, modificationParams)
						if err != nil {
							return fmt.Errorf("failed to schedule booking modification emails: %w", err)
						}
					}
				}
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update series occurrences: %w", err)
		}
	}

	return nil
}
