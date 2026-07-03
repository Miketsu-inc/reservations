package booking

import (
	"testing"
	"time"

	"github.com/bojanz/currency"
	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/stretchr/testify/assert"
	"github.com/teambition/rrule-go"
)

func ct(year int, month time.Month, day int, timeStr string, loc *time.Location) time.Time {
	t, _ := time.Parse("15:04", timeStr)
	return time.Date(year, month, day, t.Hour(), t.Minute(), 0, 0, loc)
}

func TestBuildOccurrenceTimestampUpdate(t *testing.T) {

	tz, _ := time.LoadLocation("Europe/Budapest")

	duration := time.Duration(10) * time.Minute

	servicePhaseId := 0
	seriesPhases := []domain.BookingSeriesPhase{{
		ServicePhaseId: &servicePhaseId,
		Duration:       int(duration.Minutes()),
		PhaseType:      types.ServicePhaseTypeActive,
	}}

	t.Run("1 hour offset", func(t *testing.T) {

		rrule, _ := rrule.NewRRule(rrule.ROption{
			Freq:      rrule.WEEKLY,
			Dtstart:   ct(2026, time.June, 7, "16:00", tz),
			Interval:  1,
			Byweekday: []rrule.Weekday{},
			Until:     ct(2026, time.July, 11, "16:00", tz),
		})

		context := occurrenceTimestampUpdateContext{
			rrule:                    rrule,
			merchantTz:               tz,
			duration:                 duration,
			seriesOriginalDateOffset: time.Duration(1) * time.Hour,
			seriesVersion:            3,
		}

		lastOccurrenceDate := ct(2026, time.June, 7, "14:00", time.UTC)

		fromDates := []time.Time{
			ct(2026, time.June, 14, "13:00", time.UTC),
			ct(2026, time.June, 21, "13:00", time.UTC),
			ct(2026, time.June, 28, "13:00", time.UTC),
			ct(2026, time.July, 5, "13:00", time.UTC),
		}

		seriesVersion := 1

		bookings := make([]domain.Booking, len(fromDates))
		for i, d := range fromDates {
			bookings[i] = domain.Booking{
				Id:                 i,
				Status:             types.BookingStatusConfirmed,
				FromDate:           d,
				ToDate:             d.Add(duration),
				SeriesOriginalDate: &d,
				SeriesVersion:      &seriesVersion,
			}
		}

		bookings[0].Status = types.BookingStatusCompleted

		seriesVersionDiff := 3
		bookings[2].SeriesVersion = &seriesVersionDiff

		expectedBookingIds := []int{1, 3}
		expectedFromDates := []time.Time{
			ct(2026, time.June, 21, "14:00", time.UTC),
			ct(2026, time.July, 5, "14:00", time.UTC),
		}
		expectedToDates := []time.Time{
			ct(2026, time.June, 21, "14:10", time.UTC),
			ct(2026, time.July, 5, "14:10", time.UTC),
		}
		expectedBookingPhases := []domain.BookingPhase{
			{
				BookingId:      1,
				ServicePhaseId: &servicePhaseId,
				FromDate:       expectedFromDates[0],
				ToDate:         expectedToDates[0],
				PhaseType:      types.ServicePhaseTypeActive,
			},
			{
				BookingId:      3,
				ServicePhaseId: &servicePhaseId,
				FromDate:       expectedFromDates[1],
				ToDate:         expectedToDates[1],
				PhaseType:      types.ServicePhaseTypeActive,
			},
		}

		update, err := buildOccurrenceTimestampUpdate(context, bookings, seriesPhases, lastOccurrenceDate)
		if assert.NoError(t, err, "'buildOccurrenceTimestampUpdate' should not error") {
			assert.Equal(t, expectedBookingIds, update.BookingIds, "booking ids shall match")
			assert.Equal(t, expectedFromDates, update.FromDates, "from dates shall match")
			assert.Equal(t, expectedToDates, update.ToDates, "to dates shall match")
			assert.Equal(t, expectedBookingPhases, update.BookingPhases, "booking phases shall match")
		}
	})

	t.Run("-1 hour offset", func(t *testing.T) {

		rrule, _ := rrule.NewRRule(rrule.ROption{
			Freq:      rrule.WEEKLY,
			Dtstart:   ct(2026, time.June, 7, "16:00", tz),
			Interval:  1,
			Byweekday: []rrule.Weekday{},
			Until:     ct(2026, time.July, 4, "16:00", tz),
		})

		context := occurrenceTimestampUpdateContext{
			rrule:                    rrule,
			merchantTz:               tz,
			duration:                 duration,
			seriesOriginalDateOffset: time.Duration(-1) * time.Hour,
			seriesVersion:            2,
		}

		lastOccurrenceDate := ct(2026, time.June, 7, "14:00", time.UTC)

		fromDates := []time.Time{
			ct(2026, time.June, 14, "15:00", time.UTC),
			ct(2026, time.June, 21, "15:00", time.UTC),
			ct(2026, time.June, 28, "15:00", time.UTC),
		}

		seriesVersion := 1

		bookings := make([]domain.Booking, len(fromDates))
		for i, d := range fromDates {
			bookings[i] = domain.Booking{
				Id:                 i,
				Status:             types.BookingStatusConfirmed,
				FromDate:           d,
				ToDate:             d.Add(duration),
				SeriesOriginalDate: &d,
				SeriesVersion:      &seriesVersion,
			}
		}

		bookings[0].FromDate = ct(2026, time.June, 14, "18:00", time.UTC)

		expectedBookingIds := []int{0, 1, 2}
		expectedFromDates := []time.Time{
			ct(2026, time.June, 14, "14:00", time.UTC),
			ct(2026, time.June, 21, "14:00", time.UTC),
			ct(2026, time.June, 28, "14:00", time.UTC),
		}
		expectedToDates := []time.Time{
			ct(2026, time.June, 14, "14:10", time.UTC),
			ct(2026, time.June, 21, "14:10", time.UTC),
			ct(2026, time.June, 28, "14:10", time.UTC),
		}
		expectedBookingPhases := []domain.BookingPhase{
			{
				BookingId:      0,
				ServicePhaseId: &servicePhaseId,
				FromDate:       expectedFromDates[0],
				ToDate:         expectedToDates[0],
				PhaseType:      types.ServicePhaseTypeActive,
			},
			{
				BookingId:      1,
				ServicePhaseId: &servicePhaseId,
				FromDate:       expectedFromDates[1],
				ToDate:         expectedToDates[1],
				PhaseType:      types.ServicePhaseTypeActive,
			},
			{
				BookingId:      2,
				ServicePhaseId: &servicePhaseId,
				FromDate:       expectedFromDates[2],
				ToDate:         expectedToDates[2],
				PhaseType:      types.ServicePhaseTypeActive,
			},
		}

		update, err := buildOccurrenceTimestampUpdate(context, bookings, seriesPhases, lastOccurrenceDate)
		if assert.NoError(t, err, "'buildOccurrenceTimestampUpdate' should not error") {
			assert.Equal(t, expectedBookingIds, update.BookingIds, "booking ids shall match")
			assert.Equal(t, expectedFromDates, update.FromDates, "from dates shall match")
			assert.Equal(t, expectedToDates, update.ToDates, "to dates shall match")
			assert.Equal(t, expectedBookingPhases, update.BookingPhases, "booking phases shall match")
		}
	})

	t.Run("Daylight savings", func(t *testing.T) {

		rrule, _ := rrule.NewRRule(rrule.ROption{
			Freq:      rrule.WEEKLY,
			Dtstart:   ct(2026, time.March, 15, "01:00", tz),
			Interval:  1,
			Byweekday: []rrule.Weekday{},
			Until:     ct(2026, time.April, 5, "01:00", tz),
		})

		context := occurrenceTimestampUpdateContext{
			rrule:                    rrule,
			merchantTz:               tz,
			duration:                 duration,
			seriesOriginalDateOffset: time.Duration(2) * time.Hour,
			seriesVersion:            2,
		}

		lastOccurrenceDate := ct(2026, time.March, 15, "00:00", time.UTC)

		fromDates := []time.Time{
			ct(2026, time.March, 21, "22:00", time.UTC),
			ct(2026, time.March, 28, "22:00", time.UTC),
			ct(2026, time.April, 4, "23:00", time.UTC),
		}

		seriesVersion := 1

		bookings := make([]domain.Booking, len(fromDates))
		for i, d := range fromDates {
			bookings[i] = domain.Booking{
				Id:                 i,
				Status:             types.BookingStatusConfirmed,
				FromDate:           d,
				ToDate:             d.Add(duration),
				SeriesOriginalDate: &d,
				SeriesVersion:      &seriesVersion,
			}
		}

		expectedBookingIds := []int{0, 1, 2}
		expectedFromDates := []time.Time{
			ct(2026, time.March, 22, "00:00", time.UTC),
			ct(2026, time.March, 29, "00:00", time.UTC),
			ct(2026, time.April, 4, "23:00", time.UTC),
		}
		expectedToDates := []time.Time{
			ct(2026, time.March, 22, "00:10", time.UTC),
			ct(2026, time.March, 29, "00:10", time.UTC),
			ct(2026, time.April, 4, "23:10", time.UTC),
		}
		expectedBookingPhases := []domain.BookingPhase{
			{
				BookingId:      0,
				ServicePhaseId: &servicePhaseId,
				FromDate:       expectedFromDates[0],
				ToDate:         expectedToDates[0],
				PhaseType:      types.ServicePhaseTypeActive,
			},
			{
				BookingId:      1,
				ServicePhaseId: &servicePhaseId,
				FromDate:       expectedFromDates[1],
				ToDate:         expectedToDates[1],
				PhaseType:      types.ServicePhaseTypeActive,
			},
			{
				BookingId:      2,
				ServicePhaseId: &servicePhaseId,
				FromDate:       expectedFromDates[2],
				ToDate:         expectedToDates[2],
				PhaseType:      types.ServicePhaseTypeActive,
			},
		}

		update, err := buildOccurrenceTimestampUpdate(context, bookings, seriesPhases, lastOccurrenceDate)
		if assert.NoError(t, err, "'buildOccurrenceTimestampUpdate' should not error") {
			assert.Equal(t, expectedBookingIds, update.BookingIds, "booking ids shall match")
			assert.Equal(t, expectedFromDates, update.FromDates, "from dates shall match")
			assert.Equal(t, expectedToDates, update.ToDates, "to dates shall match")
			assert.Equal(t, expectedBookingPhases, update.BookingPhases, "booking phases shall match")
		}
	})
}

func TestMakeExistingParticipantsMap(t *testing.T) {

	customer1, _ := uuid.Parse("019e9c52-3461-7da1-b84c-a8272ffaa6f9")
	customer2, _ := uuid.Parse("019e9c52-3461-7e0b-bf2e-97530a663369")

	t.Run("Basic", func(t *testing.T) {
		bookingIds := []int{0, 1, 2, 3}
		customerIdsByBooking := map[int][]uuid.UUID{
			0: {customer1},
			1: {customer2},
			3: {customer1, customer2},
		}

		expectedExistingMap := map[int]map[uuid.UUID]struct{}{
			0: {
				customer1: {},
			},
			1: {
				customer2: {},
			},
			2: {},
			3: {
				customer1: {},
				customer2: {},
			},
		}

		existingMap := makeExistingParticipantsMap(bookingIds, customerIdsByBooking)
		assert.Equal(t, expectedExistingMap, existingMap, "existing participant maps shall match")
	})
}

func TestCalculateCapacity(t *testing.T) {

	existingParticipant1, _ := uuid.Parse("019e9c52-3461-7da1-b84c-a8272ffaa6f9")
	existingParticipant2, _ := uuid.Parse("019e9c52-3461-7e0b-bf2e-97530a663369")

	futureBookingsMap := map[int]domain.Booking{
		0: {
			Id:                  0,
			CurrentParticipants: 0,
			MaxParticipants:     4,
		},
		1: {
			Id:                  1,
			CurrentParticipants: 1,
			MaxParticipants:     5,
		},
		2: {
			Id:                  2,
			CurrentParticipants: 2,
			MaxParticipants:     3,
		},
	}
	existingParticipantsByBooking := map[int]map[uuid.UUID]struct{}{
		0: {},
		1: {
			existingParticipant1: {},
		},
		2: {
			existingParticipant1: {},
			existingParticipant2: {},
		},
	}

	t.Run("Insert participants", func(t *testing.T) {

		toDeleteMap := make(map[uuid.UUID]struct{})

		participantToInsert1, _ := uuid.Parse("019e9c52-3461-7b73-b927-603a106177a1")
		participantToInsert2, _ := uuid.Parse("019e9c52-3461-78e9-bfe9-0413179e1230")

		toInsertMap := map[uuid.UUID]struct{}{
			existingParticipant1: {},
			participantToInsert1: {},
			participantToInsert2: {},
		}

		capacity := calculateCapacity(futureBookingsMap, existingParticipantsByBooking, toDeleteMap, toInsertMap)

		assert.Equal(t, []int{0, 1}, capacity.BookingIdsToUpdate, "booking ids to update shall match")
		assert.Equal(t, []int{3, 2}, capacity.DeltaToInsert, "capacity delta to insert shall match")
		assert.ElementsMatch(t, []int{2}, capacity.BookingIdsExceeded, "booking ids exceeding max participants shall match")
		assert.Equal(t, map[int]int{0: 3, 1: 3, 2: 2}, capacity.ByBooking, "capacity delta by booking shall match")
	})

	t.Run("Delete participants", func(t *testing.T) {

		participantToDelte, _ := uuid.Parse("019e9c52-3461-75be-a296-818f5d59cc47")

		toDeleteMap := map[uuid.UUID]struct{}{
			existingParticipant1: {},
			participantToDelte:   {},
		}

		toInsertMap := make(map[uuid.UUID]struct{})

		capacity := calculateCapacity(futureBookingsMap, existingParticipantsByBooking, toDeleteMap, toInsertMap)

		assert.Equal(t, []int{1, 2}, capacity.BookingIdsToUpdate, "booking ids to update shall match")
		assert.Equal(t, []int{-1, -1}, capacity.DeltaToInsert, "capacity delta to insert shall match")
		assert.ElementsMatch(t, []int{}, capacity.BookingIdsExceeded, "booking ids exceeding max participants shall match")
		assert.Equal(t, map[int]int{0: 0, 1: 0, 2: 1}, capacity.ByBooking, "capacity delta by booking shall match")
	})

	t.Run("Participants unchanged", func(t *testing.T) {
		toDeleteMap := make(map[uuid.UUID]struct{})
		toInsertMap := make(map[uuid.UUID]struct{})

		capacity := calculateCapacity(futureBookingsMap, existingParticipantsByBooking, toDeleteMap, toInsertMap)

		assert.Equal(t, []int{}, capacity.BookingIdsToUpdate, "booking ids to update shall match")
		assert.Equal(t, []int{}, capacity.DeltaToInsert, "capacity delta to insert shall match")
		assert.ElementsMatch(t, []int{}, capacity.BookingIdsExceeded, "booking ids exceeding max participants shall match")
		assert.Equal(t, map[int]int{0: 0, 1: 1, 2: 2}, capacity.ByBooking, "capacity delta by booking shall match")
	})
}

func TestCheckCapacityUpdateSuccess(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		toUpdate := []int{0, 1, 2, 3, 4, 5}
		updated := []int{0, 2, 4, 5}

		expectedFailedToUpdate := []int{1, 3}

		failedToUpdate := checkCapacityUpdateSuccess(toUpdate, updated)
		assert.Equal(t, expectedFailedToUpdate, failedToUpdate, "failed to updates shall match")
	})
}

func TestBuildOccurrenceParticipantsToInsert(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		participant1, _ := uuid.Parse("019e9c52-3461-7da1-b84c-a8272ffaa6f9")
		participant2, _ := uuid.Parse("019e9c52-3461-7e0b-bf2e-97530a663369")

		bookingIds := []int{0, 1, 2}
		requsetedToInsert := []uuid.UUID{participant1}
		existingByBooking := map[int]map[uuid.UUID]struct{}{
			0: {
				participant1: {},
			},
			1: {
				participant2: {},
			},
			2: {},
			3: {
				participant2: {},
			},
		}

		expectedParticipantsToInsert := []domain.BookingParticipant{
			{
				BookingId:  1,
				CustomerId: &participant1,
				Status:     types.BookingStatusConfirmed,
			},
			{
				BookingId:  2,
				CustomerId: &participant1,
				Status:     types.BookingStatusConfirmed,
			},
		}

		participantsToInsert := buildOccurrenceParticipantsToInsert(bookingIds, requsetedToInsert, existingByBooking)
		assert.Equal(t, expectedParticipantsToInsert, participantsToInsert, "booking participants to insert shall match")
	})
}

func TestCalculateTotalPrices(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		bookingIds := []int{0, 1, 2}
		capacityByBooking := map[int]int{
			bookingIds[0]: 0,
			bookingIds[1]: 1,
			bookingIds[2]: 2,
		}

		pricePerPerson, _ := currency.NewAmount("1000", "HUF")

		totalPrice1, _ := currency.NewAmount("0", "HUF")
		totalPrice2 := pricePerPerson
		totalPrice3, _ := currency.NewAmount("2000", "HUF")

		expectedTotalPrices := []currencyx.Price{
			{Amount: totalPrice1},
			{Amount: totalPrice2},
			{Amount: totalPrice3},
		}

		totalPrices, err := calculateTotalPrices(currencyx.Price{Amount: pricePerPerson}, bookingIds, capacityByBooking)
		if assert.NoError(t, err, "'calculateTotalPrices' should not error") {
			assert.Equal(t, expectedTotalPrices, totalPrices, "total prices shall match")
		}
	})
}
