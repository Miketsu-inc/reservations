package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type bookingRepository struct {
	db db.DBTX
}

func NewBookingRepository(db db.DBTX) domain.BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) WithTx(tx db.DBTX) domain.BookingRepository {
	return &bookingRepository{db: tx}
}

func (r *bookingRepository) NewBooking(ctx context.Context, booking domain.Booking) (int, error) {
	query := `
	insert into "Booking" (status, booking_type, is_recurring, merchant_id, employee_id, service_id, location_id, booking_series_id, series_original_date, from_date, to_date,
		service_name, price_per_person, total_price, price_type, formatted_location, merchant_note, min_participants, max_participants, current_participants,
		occurrence_index, series_version)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	returning id
	`

	var bookingId int

	err := r.db.QueryRow(ctx, query, booking.Status, booking.BookingType, booking.IsRecurring, booking.MerchantId, booking.EmployeeId, booking.ServiceId, booking.LocationId,
		booking.BookingSeriesId, booking.SeriesOriginalDate, booking.FromDate, booking.ToDate, booking.ServiceName, booking.PricePerPerson, booking.TotalPrice,
		booking.PriceType, booking.FormattedLocation, booking.MerchantNote, booking.MinParticipants, booking.MaxParticipants, booking.CurrentParticipants,
		booking.OccurrenceIndex, booking.SeriesVersion).Scan(&bookingId)
	if err != nil {
		return 0, fmt.Errorf("NewBooking: %w", err)
	}

	return bookingId, nil
}

func (r *bookingRepository) NewBookings(ctx context.Context, bookings []domain.Booking) ([]int, error) {
	query := `
	insert into "Booking" (status, booking_type, is_recurring, merchant_id, employee_id, service_id, location_id, booking_series_id, series_original_date, from_date, to_date,
		service_name, price_per_person, total_price, price_type, formatted_location, merchant_note, min_participants, max_participants, current_participants, occurrence_index, series_version)
	select unnest($1::text[]), unnest($2::text[]), unnest($3::boolean[]), unnest($4::uuid[]), unnest($5::int[]), unnest($6::int[]), unnest($7::int[]),
		unnest($8::int[]), unnest($9::timestamptz[]), unnest($10::timestamptz[]), unnest($11::timestamptz[]), unnest($12::text[]), unnest($13::price[]),
		unnest($14::price[]), unnest($15::text[]), unnest($16::text[]), unnest($17::text[]), unnest($18::int[]), unnest($19::int[]), unnest($20::int[]),
		unnest($21::int[]), unnest($22::int[])
	returning id
	`

	var bookingIds []int

	bookingsCount := len(bookings)

	statuses := make([]string, bookingsCount)
	types := make([]string, bookingsCount)
	isRecurrings := make([]bool, bookingsCount)
	merchantIds := make([]uuid.UUID, bookingsCount)
	employeeIds := make([]pgtype.Int4, bookingsCount)
	serviceIds := make([]pgtype.Int4, bookingsCount)
	locationIds := make([]int, bookingsCount)
	seriesIds := make([]pgtype.Int4, bookingsCount)
	seriesOriginalDates := make([]pgtype.Timestamptz, bookingsCount)
	fromDates := make([]time.Time, bookingsCount)
	toDates := make([]time.Time, bookingsCount)
	serviceNames := make([]string, bookingsCount)
	pricePerPersons := make([]currencyx.Price, bookingsCount)
	totalPrices := make([]currencyx.Price, bookingsCount)
	priceTypes := make([]string, bookingsCount)
	formattedLocations := make([]string, bookingsCount)
	merchantNotes := make([]pgtype.Text, bookingsCount)
	minParicipants := make([]int, bookingsCount)
	maxParicipants := make([]int, bookingsCount)
	currentParicipants := make([]int, bookingsCount)
	occurrenceIndexes := make([]pgtype.Int4, bookingsCount)
	seriesVersions := make([]pgtype.Int4, bookingsCount)

	for i, b := range bookings {
		statuses[i] = b.Status.String()
		types[i] = b.BookingType.String()
		isRecurrings[i] = b.IsRecurring
		merchantIds[i] = b.MerchantId
		if b.EmployeeId == nil {
			employeeIds[i] = pgtype.Int4{Valid: false}
		} else {
			employeeIds[i] = pgtype.Int4{Int32: int32(*b.EmployeeId), Valid: true}
		}
		if b.ServiceId == nil {
			serviceIds[i] = pgtype.Int4{Valid: false}
		} else {
			serviceIds[i] = pgtype.Int4{Int32: int32(*b.ServiceId), Valid: true}
		}
		locationIds[i] = b.LocationId
		if b.BookingSeriesId == nil {
			seriesIds[i] = pgtype.Int4{Valid: false}
		} else {
			seriesIds[i] = pgtype.Int4{Int32: int32(*b.BookingSeriesId), Valid: true}
		}
		if b.IsRecurring {
			seriesOriginalDates[i] = pgtype.Timestamptz{Time: b.FromDate, Valid: true}
		} else {
			seriesOriginalDates[i] = pgtype.Timestamptz{Valid: false}
		}
		fromDates[i] = b.FromDate
		toDates[i] = b.ToDate
		serviceNames[i] = b.ServiceName
		pricePerPersons[i] = b.PricePerPerson
		totalPrices[i] = b.TotalPrice
		priceTypes[i] = b.PriceType.String()
		formattedLocations[i] = b.FormattedLocation
		if b.MerchantNote == nil {
			merchantNotes[i] = pgtype.Text{Valid: false}
		} else {
			merchantNotes[i] = pgtype.Text{String: *b.MerchantNote, Valid: true}
		}
		minParicipants[i] = b.MinParticipants
		maxParicipants[i] = b.MaxParticipants
		currentParicipants[i] = b.CurrentParticipants
		if b.OccurrenceIndex == nil {
			occurrenceIndexes[i] = pgtype.Int4{Valid: false}
		} else {
			occurrenceIndexes[i] = pgtype.Int4{Int32: int32(*b.OccurrenceIndex), Valid: true}
		}
		if b.SeriesVersion == nil {
			seriesVersions[i] = pgtype.Int4{Valid: false}
		} else {
			seriesVersions[i] = pgtype.Int4{Int32: int32(*b.SeriesVersion), Valid: true}
		}
	}

	rows, _ := r.db.Query(ctx, query, statuses, types, isRecurrings, merchantIds, employeeIds, serviceIds, locationIds, seriesIds, seriesOriginalDates,
		fromDates, toDates, serviceNames, pricePerPersons, totalPrices, priceTypes, formattedLocations, merchantNotes, minParicipants, maxParicipants, currentParicipants,
		occurrenceIndexes, seriesVersions)
	bookingIds, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return []int{}, fmt.Errorf("NewBookings: %w", err)
	}

	return bookingIds, nil
}

func (r *bookingRepository) NewBookingPhases(ctx context.Context, bookingPhases []domain.BookingPhase) error {
	query := `
	insert into "BookingPhase" (booking_id, service_phase_id, from_date, to_date, phase_type)
	select unnest($1::int[]), unnest($2::int[]), unnest($3::timestamptz[]), unnest($4::timestamptz[]), unnest($5::text[])
	`

	bookingIds := make([]int, len(bookingPhases))
	servicePhaseIds := make([]pgtype.Int4, len(bookingPhases))
	fromDates := make([]time.Time, len(bookingPhases))
	toDates := make([]time.Time, len(bookingPhases))
	phaseTypes := make([]string, len(bookingPhases))

	for i, bp := range bookingPhases {
		bookingIds[i] = bp.BookingId
		if bp.ServicePhaseId == nil {
			servicePhaseIds[i] = pgtype.Int4{Valid: false}
		} else {
			servicePhaseIds[i] = pgtype.Int4{Int32: int32(*bp.ServicePhaseId), Valid: true}
		}
		fromDates[i] = bp.FromDate
		toDates[i] = bp.ToDate
		phaseTypes[i] = bp.PhaseType.String()
	}

	_, err := r.db.Exec(ctx, query, bookingIds, servicePhaseIds, fromDates, toDates, phaseTypes)
	if err != nil {
		return fmt.Errorf("NewBookingPhases: %w", err)
	}

	return nil
}

func (r *bookingRepository) NewBookingParticipants(ctx context.Context, bookingParticipants []domain.BookingParticipant) error {
	query := `
	insert into "BookingParticipant" (status, booking_id, customer_id, customer_note)
	select unnest($1::text[]), unnest($2::int[]), unnest($3::uuid[]), unnest($4::text[])
	`

	statuses := make([]string, len(bookingParticipants))
	bookingIds := make([]int, len(bookingParticipants))
	customerIds := make([]pgtype.UUID, len(bookingParticipants))
	customerNotes := make([]pgtype.Text, len(bookingParticipants))

	for i, bp := range bookingParticipants {
		statuses[i] = bp.Status.String()
		bookingIds[i] = bp.BookingId
		if bp.CustomerId == nil {
			customerIds[i] = pgtype.UUID{Valid: false}
		} else {
			customerIds[i] = pgtype.UUID{Bytes: *bp.CustomerId, Valid: true}
		}
		if bp.CustomerNote == nil {
			customerNotes[i] = pgtype.Text{Valid: false}
		} else {
			customerNotes[i] = pgtype.Text{String: *bp.CustomerNote, Valid: true}
		}
	}

	_, err := r.db.Exec(ctx, query, statuses, bookingIds, customerIds, customerNotes)
	if err != nil {
		return fmt.Errorf("NewBookingParticipants: %w", err)
	}

	return nil
}

func (r *bookingRepository) DeleteBookingPhasesBatch(ctx context.Context, bookingIds []int) error {
	query := `delete from "BookingPhase" where booking_id = any($1::int[])`

	_, err := r.db.Exec(ctx, query, bookingIds)
	if err != nil {
		return fmt.Errorf("DeleteBookingPhasesBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) DeleteBookingParticipantsBatch(ctx context.Context, bookingIds []int, participantIds []uuid.UUID) error {
	query := `delete from "BookingParticipant"
	where booking_id = any($1::int[]) and customer_id = any($2::uuid[])`

	_, err := r.db.Exec(ctx, query, bookingIds, participantIds)
	if err != nil {
		return fmt.Errorf("DeleteBookingParticipantsBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingStatus(ctx context.Context, merchantId uuid.UUID, bookingId int, status types.BookingStatus) error {
	query := `
	update "Booking"
	set status = $3
	where id = $1 and merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, bookingId, merchantId, status)
	if err != nil {
		return fmt.Errorf("UpdateBookingStatus: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingCoreBatch(ctx context.Context, merchantId uuid.UUID, bookingIds []int, serviceId *int, employeeId *int,
	fromDates []time.Time, toDates []time.Time, bookingType types.BookingType, status types.BookingStatus, merchantNote *string) error {
	query := `
	update "Booking" as b
	set from_date = data.new_from_dates,
	    to_date = data.new_to_dates,
	    service_id = $2,
	    booking_type = $3,
	    status = $4,
		merchant_note = $8,
		employee_id = $9
	from (select unnest($1::int[]) as id, unnest($5::timestamptz[]) as new_from_dates, unnest($6::timestamptz[]) as new_to_dates) as data
	where b.id = data.id and b.merchant_id = $7 and b.status not in ('cancelled', 'completed')
	`

	_, err := r.db.Exec(ctx, query, bookingIds, serviceId, bookingType, status, fromDates, toDates, merchantId, merchantNote, employeeId)
	if err != nil {
		return fmt.Errorf("UpdateBookingCoreBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingSeriesOriginalDateAndVersion(ctx context.Context, bookingId int, seriesOriginalDate time.Time, seriesVersion int) error {
	query := `
	update "Booking"
	set series_original_date = $2, series_version = $3
	where id = $1 and status not in ('cancelled', 'completed', 'no-show')
	`

	_, err := r.db.Exec(ctx, query, bookingId, seriesOriginalDate, seriesVersion)
	if err != nil {
		return fmt.Errorf("UpdateBookingSeriesOriginalDateAndVersion: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingPricePerPersonBatch(ctx context.Context, bookingIds []int, price currencyx.Price) error {
	query := `
	update "Booking"
	set price_per_person = $2
	where id = any($1::int[]) and status not in ('cancelled', 'completed', 'no-show')
	`

	_, err := r.db.Exec(ctx, query, bookingIds, price)
	if err != nil {
		return fmt.Errorf("UpdateBookingPricePerPersonBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingTotalPriceBatch(ctx context.Context, bookingIds []int, prices []currencyx.Price) error {
	query := `
	update "Booking" b
	set total_price = u.total_price
	from (
		select unnest($1::int[]) as id, unnest($2::price[]) as total_price
		) as u
	where b.id = u.id and b.status not in ('cancelled', 'completed', 'no-show')
	`

	_, err := r.db.Exec(ctx, query, bookingIds, prices)
	if err != nil {
		return fmt.Errorf("UpdateBookingTotalPriceBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingDetailsBatch(ctx context.Context, merchantId uuid.UUID, bookingIds []int, details []domain.BookingDetails) error {
	query := `
	update "Booking" b
	set price_per_person = u.price_per_person, total_price = u.total_price, min_participants = u.min_participants,
		max_participants = u.max_participants, current_participants = u.current_participants
	from (
		select unnest($2::int[]) as id, unnest($3::price[]) as price_per_person, unnest($4::price[]) as total_price,
			unnest($5::int[]) as min_participants, unnest($6::int[]) as max_participants, unnest($7::int[]) as current_participants
		) as u
	where b.id = u.id and b.merchant_id = $1 and b.status not in ('cancelled', 'completed', 'no-show')
	`

	pricePerPersons := make([]currencyx.Price, len(details))
	totalPrices := make([]currencyx.Price, len(details))
	minParticipants := make([]int, len(details))
	maxParicipants := make([]int, len(details))
	currentParicipants := make([]int, len(details))

	for i, d := range details {
		pricePerPersons[i] = d.PricePerPerson
		totalPrices[i] = d.TotalPrice
		minParticipants[i] = d.MinParticipants
		maxParicipants[i] = d.MaxParticipants
		currentParicipants[i] = d.CurrentParticipants
	}

	_, err := r.db.Exec(ctx, query, merchantId, bookingIds, pricePerPersons, totalPrices,
		minParticipants, maxParicipants, currentParicipants)
	if err != nil {
		return fmt.Errorf("UpdateBookingDetailsBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingEmployeeBatch(ctx context.Context, bookingIds []int, employeeId *int) error {
	query := `
	update "Booking"
	set employee_id = $2
	where id = any($1::int[]) and status not in ('cancelled', 'completed', 'no-show')
	`

	_, err := r.db.Exec(ctx, query, bookingIds, employeeId)
	if err != nil {
		return fmt.Errorf("UpdateBookingEmployeeBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingOccurrencesBatch(ctx context.Context, bookingIds []int, fromDates, toDates []time.Time, seriesId int, seriesVersion int) error {
	query := `
	update "Booking" b
	set from_date = u.from_date, to_date = u.to_date, booking_series_id = $4, series_original_date = u.from_date, series_version = $5
	from unnest($1::int[], $2::timestamptz[], $3::timestamptz[])
		as u(id, from_date, to_date)
	where b.id = u.id and b.from_date > now() and b.status not in ('cancelled', 'completed', 'no-show')
	`

	_, err := r.db.Exec(ctx, query, bookingIds, fromDates, toDates, seriesId, seriesVersion)
	if err != nil {
		return fmt.Errorf("UpdateBookingOccurrencesBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingParticipants(ctx context.Context, participants []domain.BookingParticipant, updateStatusOnConflict bool) error {
	query := `
	insert into "BookingParticipant" (booking_id, customer_id, status)
	select unnest($1::int[]), unnest($2::uuid[]), unnest($3::text[])
	on conflict (booking_id, customer_id)
	do update
	set cancelled_on = NULL, cancellation_reason = NULL,
		status = case
			when $4 then excluded.status
			else "BookingParticipant".status
		end
	`

	statuses := make([]string, len(participants))
	bookingIds := make([]int, len(participants))
	customerIds := make([]pgtype.UUID, len(participants))

	for i, bp := range participants {
		statuses[i] = bp.Status.String()
		bookingIds[i] = bp.BookingId
		if bp.CustomerId == nil {
			customerIds[i] = pgtype.UUID{Valid: false}
		} else {
			customerIds[i] = pgtype.UUID{Bytes: *bp.CustomerId, Valid: true}
		}
	}

	_, err := r.db.Exec(ctx, query, bookingIds, customerIds, statuses, updateStatusOnConflict)
	if err != nil {
		return fmt.Errorf("UpdateBookingParticipants: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateParticipantStatus(ctx context.Context, bookingId int, participantId int, status types.BookingStatus) error {
	query := `
	update "BookingParticipant"
	set status = $3,
		cancelled_on = case
			when $3 = 'cancelled'
			then coalesce(cancelled_on, now())
			else cancelled_on
		end
	where booking_id = $1 and id = $2
	`

	_, err := r.db.Exec(ctx, query, bookingId, participantId, status)
	if err != nil {
		return fmt.Errorf("UpdateParticipantStatus: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateParticipantCountBatch(ctx context.Context, bookingIds []int, participantDelta []int) ([]int, error) {
	assert.True(len(bookingIds) == len(participantDelta), "booking ids and participant delta length should be the same", len(bookingIds), len(participantDelta))

	query := `
	update "Booking" b
	set current_participants = b.current_participants + u.delta
	from unnest($1::int[], $2::int[]) as u(id, delta)
	where b.id = u.id and b.booking_type in ('event', 'class') and b.status not in ('cancelled', 'completed')
		and b.current_participants + u.delta <= b.max_participants and b.current_participants + u.delta > 0
	returning b.id
	`

	rows, _ := r.db.Query(ctx, query, bookingIds, participantDelta)
	bookingIds, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return []int{}, fmt.Errorf("UpdateParticipantCountBatch: %w", err)
	}

	return bookingIds, nil
}

func (r *bookingRepository) DecrementEveryParticipantCountForCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	update "Booking" b
	set current_participants = current_participants - 1
	from "BookingParticipant" bp
	where b.id = bp.booking_id and bp.customer_id = $1 and b.merchant_id = $2 and b.booking_type in ('event', 'class')
	`

	_, err := r.db.Exec(ctx, query, customerId, merchantId)
	if err != nil {
		return fmt.Errorf("DecrementEveryParticipantCountForCustomer: %w", err)
	}

	return nil
}

func (r *bookingRepository) TransferDummyBookings(ctx context.Context, merchantId uuid.UUID, fromCustomer uuid.UUID, toCustomer uuid.UUID) error {
	query := `
	update "BookingParticipant" bp
	set transferred_to = $3
	from "Booking" b
	join "Customer" c on bp.customer_id = c.id
	where b.merchant_id = $1 and bp.booking_id = b.id and bp.customer_id = $2 and c.user_id is null
	`

	_, err := r.db.Exec(ctx, query, merchantId, fromCustomer, toCustomer)
	if err != nil {
		return fmt.Errorf("TransferDummyBookings: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingPhasesBatch(ctx context.Context, bookingPhases []domain.BookingPhase) error {
	query := `
	update "BookingPhase" bp
	set booking_id = u.booking_id, service_phase_id = u.service_phase_id, from_date = u.from_date, to_date = u.to_date, phase_type = u.phase_type
	from unnest($1::int[], $2::int[], $3::int[], $4::timestamptz[], $5::timestamptz[], $6::text[])
		as u(id, booking_id, service_phase_id, from_date, to_date, phase_type)
	where bp.id = u.id
	`

	bookingPhasesCount := len(bookingPhases)

	ids := make([]int, bookingPhasesCount)
	bookingIds := make([]int, bookingPhasesCount)
	servicePhaseIds := make([]pgtype.Int4, bookingPhasesCount)
	fromDates := make([]time.Time, bookingPhasesCount)
	toDates := make([]time.Time, bookingPhasesCount)
	phaseTypes := make([]string, bookingPhasesCount)

	for i, bp := range bookingPhases {
		ids[i] = bp.Id
		bookingIds[i] = bp.BookingId
		if bp.ServicePhaseId == nil {
			servicePhaseIds[i] = pgtype.Int4{Valid: false}
		} else {
			servicePhaseIds[i] = pgtype.Int4{Int32: int32(*bp.ServicePhaseId), Valid: true}
		}
		fromDates[i] = bp.FromDate
		toDates[i] = bp.ToDate
		phaseTypes[i] = bp.PhaseType.String()
	}

	_, err := r.db.Exec(ctx, query, ids, bookingIds, servicePhaseIds, fromDates, toDates, phaseTypes)
	if err != nil {
		return fmt.Errorf("UpdateBookingPhasesBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) CancelBookingByMerchant(ctx context.Context, merchantId uuid.UUID, bookingId int, cancellationReason string) error {
	query := `
	update "Booking"
	set status = 'cancelled', cancelled_by_merchant_on = $1, cancellation_reason = $2
	where id = $4 and merchant_id = $3
	`

	_, err := r.db.Exec(ctx, query, time.Now().UTC(), cancellationReason, merchantId, bookingId)
	if err != nil {
		return fmt.Errorf("CancelBookingByMerchant: %w", err)
	}

	return nil
}

func (r *bookingRepository) CancelBookingByMerchantBatch(ctx context.Context, bookingIds []int) error {
	query := `
	update "Booking"
	set status = 'cancelled', cancelled_by_merchant_on = $2
	where id = any($1::int[]) and status not in ('cancelled', 'completed', 'no-show') and from_date > now()
	`

	_, err := r.db.Exec(ctx, query, bookingIds, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("CancelBookingByMerchantBatch: %w", err)
	}

	return nil
}

func (r *bookingRepository) DeleteAppointmentsByCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	delete from "Booking" b
	using "BookingParticipant" bp
	where bp.booking_id = b.id and bp.customer_id = $1 and b.merchant_id = $2 and b.booking_type = 'appointment'
	`

	_, err := r.db.Exec(ctx, query, customerId, merchantId)
	if err != nil {
		return fmt.Errorf("DeleteAppointmentsByCustomer: %w", err)
	}

	return nil
}

func (r *bookingRepository) DeleteParticipantByCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	delete from "BookingParticipant" bp
	using "Booking" b
	where bp.booking_id = b.id and bp.customer_id = $1 and b.merchant_id = $2
	`

	_, err := r.db.Exec(ctx, query, customerId, merchantId)
	if err != nil {
		return fmt.Errorf("DeleteParticipantByCustomer: %w", err)
	}

	return nil
}

func (r *bookingRepository) GetBooking(ctx context.Context, bookingId int) (domain.Booking, error) {
	query := `
	select * from "Booking"
	where id = $1
	`

	rows, _ := r.db.Query(ctx, query, bookingId)
	booking, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.Booking])
	if err != nil {
		return domain.Booking{}, fmt.Errorf("GetBooking: %w", err)
	}

	return booking, nil
}

func (r *bookingRepository) GetPublicBooking(ctx context.Context, bookingId int, userId uuid.UUID) (domain.PublicBooking, error) {
	query := `
	select b.from_date, b.to_date, b.price_per_person as price, m.name as merchant_name, b.service_name, m.cancel_deadline, b.price_type,
		b.status, b.formatted_location
	from "BookingParticipant" bp
	join "Customer" c on c.id = bp.customer_id
	join "Booking" b on b.id = bp.booking_id
	join "Merchant" m on m.id = b.merchant_id
	where bp.booking_id = $1 and c.user_id = $2
	`

	var data domain.PublicBooking
	err := r.db.QueryRow(ctx, query, bookingId, userId).Scan(&data.FromDate, &data.ToDate, &data.Price, &data.MerchantName,
		&data.ServiceName, &data.CancelDeadline, &data.PriceType, &data.Status, &data.FormattedLocation)
	if err != nil {
		return domain.PublicBooking{}, fmt.Errorf("GetPublicBooking: %w", err)
	}

	return data, nil
}

func (r *bookingRepository) GetLatestBookings(ctx context.Context, merchantId uuid.UUID, afterDate time.Time, rowLimit int) ([]domain.PublicBookingDetails, error) {
	query := `
	select b.id, b.status, b.from_date, b.to_date, bp.customer_note, b.merchant_note, b.total_price as price, b.service_name,
		s.color as service_color,
		coalesce(c.first_name, u.first_name) as first_name,
		coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	left join "Service" s on b.service_id = s.id
	left join "BookingParticipant" bp on bp.booking_id = b.id and bp.status not in ('completed', 'cancelled')
	left join "Customer" c on bp.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.status not in ('completed', 'cancelled')
	order by b.id desc
	limit $3
	`

	rows, _ := r.db.Query(ctx, query, merchantId, afterDate, rowLimit)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.PublicBookingDetails])
	if err != nil {
		return []domain.PublicBookingDetails{}, fmt.Errorf("GetLatestBookings: %w", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetUpcomingBookings(ctx context.Context, merchantId uuid.UUID, afterDate time.Time, rowLimit int) ([]domain.PublicBookingDetails, error) {
	query := `
	select b.id, b.status, b.from_date, b.to_date, bp.customer_note, b.merchant_note, b.total_price as price, b.service_name,
		s.color as service_color,
		coalesce(c.first_name, u.first_name) as first_name,
		coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	left join "Service" s on b.service_id = s.id
	left join "BookingParticipant" bp on bp.booking_id = b.id and bp.status not in ('completed', 'cancelled')
	left join "Customer" c on bp.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.status not in ('completed', 'cancelled')
	order by b.from_date
	limit $3
	`

	rows, _ := r.db.Query(ctx, query, merchantId, afterDate, rowLimit)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.PublicBookingDetails])
	if err != nil {
		return []domain.PublicBookingDetails{}, fmt.Errorf("GetUpcomingBookings: %w", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetBookingsForCalendar(ctx context.Context, merchantId uuid.UUID, startTime, endTime string) ([]domain.BookingForCalendar, error) {
	query := `
	with participants as (
		select
			bp.booking_id,
			jsonb_agg(
				jsonb_build_object(
					'id', bp.id,
					'customer_id', c.id,
					'first_name', coalesce(c.first_name, u.first_name),
					'last_name', coalesce(c.last_name, u.last_name),
					'customer_note', bp.customer_note,
					'participant_status', bp.status
				)
			) as participants
		from "BookingParticipant" bp
		left join "Customer" c on bp.customer_id = c.id
		left join "User" u on c.user_id = u.id
		where bp.status not in ('cancelled')
		group by bp.booking_id
	)
	select b.id, b.booking_type, b.status as booking_status, b.is_recurring, b.from_date, b.to_date, b.merchant_note, b.price_per_person as price, b.price_type,
		b.employee_id, b.service_id, b.service_name, s.color as service_color, b.max_participants,
		coalesce(p.participants, '[]'::jsonb) as participants
	from "Booking" b
	left join "Service" s on b.service_id = s.id
	left join participants p on p.booking_id = b.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.to_date <= $3 AND b.status not in ('cancelled')
	order by b.id
	`

	rows, _ := r.db.Query(ctx, query, merchantId, startTime, endTime)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingForCalendar])
	if err != nil {
		return []domain.BookingForCalendar{}, fmt.Errorf("GetBookingsForCalendar: %w", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetBookingForExternalCalendar(ctx context.Context, bookingId int) (domain.BookingForExternalCalendar, error) {
	query := `
	select b.id, b.status, b.booking_type, b.employee_id, b.service_name, s.description as service_description, b.price_type,
		b.formatted_location, b.from_date, b.to_date, b.total_price, b.merchant_note, b.current_participants
	from "Booking" b
	join "Service" s on b.service_id = s.id
	where b.id = $1
	`

	rows, _ := r.db.Query(ctx, query, bookingId)
	bookingData, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingForExternalCalendar])
	if err != nil {
		return domain.BookingForExternalCalendar{}, fmt.Errorf("GetBookingForExternalCalendar: %w", err)
	}

	return bookingData, nil
}

func (r *bookingRepository) GetBookingForEmail(ctx context.Context, bookingId int, customerId uuid.UUID) (domain.BookingForEmail, error) {
	query := `
	select b.id, b.status, b.from_date, b.to_date, b.service_name, b.service_id, m.name as merchant_name, m.url_name as merchant_url, m.timezone,
		coalesce(s.cancel_deadline, m.cancel_deadline) as cancel_deadline, b.formatted_location, c.id as customer_id, coalesce(c.email, u.email) as customer_email,
		bp.status as participant_status, u.language
	from "Booking" b
	join "Merchant" m on m.id = b.merchant_id
	left join "Service" s on s.id = b.service_id
	left join "BookingParticipant" bp on bp.booking_id = b.id and bp.customer_id = $2
	left join "Customer" c on c.id = bp.customer_id
	left join "User" u on u.id = c.user_id
	where b.id = $1
	`

	rows, _ := r.db.Query(ctx, query, bookingId, customerId)
	booking, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingForEmail])
	if err != nil {
		return domain.BookingForEmail{}, fmt.Errorf("GetBookingForEmail: %w", err)
	}

	return booking, nil
}

func (r *bookingRepository) GetBookingParticipantByUser(ctx context.Context, bookingId int, userId uuid.UUID) (domain.BookingParticipant, error) {
	query := `
	select bp.*
	from "BookingParticipant" bp
	join "Customer" c on c.id = bp.customer_id
	where bp.booking_id = $1 and c.user_id = $2
	`

	rows, _ := r.db.Query(ctx, query, bookingId, userId)
	participant, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingParticipant])
	if err != nil {
		return domain.BookingParticipant{}, fmt.Errorf("GetBookingParticipantByUser: %w", err)
	}

	return participant, nil
}

func (r *bookingRepository) GetBookingParticipant(ctx context.Context, participantId int) (domain.BookingParticipant, error) {
	query := `
	select *
	from "BookingParticipant"
	where id = $1
	`

	rows, _ := r.db.Query(ctx, query, participantId)
	participant, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingParticipant])
	if err != nil {
		return domain.BookingParticipant{}, fmt.Errorf("GetBookingParticipant: %w", err)
	}

	return participant, nil
}

func (r *bookingRepository) GetBookingParticipants(ctx context.Context, bookingId int) ([]domain.BookingParticipant, error) {
	query := `
	select *
	from "BookingParticipant"
	where booking_id = $1
	`

	rows, _ := r.db.Query(ctx, query, bookingId)
	participants, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingParticipant])
	if err != nil {
		return []domain.BookingParticipant{}, fmt.Errorf("GetBookingParticipants: %w", err)
	}

	return participants, nil
}

func (r *bookingRepository) GetParticipantCustomerIdsForBookings(ctx context.Context, bookingIds []int) (map[int][]uuid.UUID, error) {
	query := `
	select booking_id, array_agg(customer_id) as customer_ids
	from "BookingParticipant"
	where booking_id = any($1::int[])
	group by booking_id
	`

	type bookingCustomers struct {
		BookingId   int         `db:"booking_id"`
		CustomerIds []uuid.UUID `db:"customer_ids"`
	}

	rows, _ := r.db.Query(ctx, query, bookingIds)
	bookingsWithCustomers, err := pgx.CollectRows(rows, pgx.RowToStructByName[bookingCustomers])
	if err != nil {
		return map[int][]uuid.UUID{}, fmt.Errorf("GetParticipantCustomerIdsForBookings: %w", err)
	}

	bookingCustomersMap := make(map[int][]uuid.UUID, len(bookingsWithCustomers))
	for _, b := range bookingsWithCustomers {
		bookingCustomersMap[b.BookingId] = b.CustomerIds
	}

	return bookingCustomersMap, nil
}

func (r *bookingRepository) GetUpcomingBookingsForUser(ctx context.Context, userId uuid.UUID, limit int, cursorStart time.Time, cursorId int) ([]domain.BookingForUser, error) {
	query := `
	select b.id, b.status, b.booking_type, b.is_recurring, b.from_date, b.to_date, b.price_per_person, m.name as merchant_name,
		m.url_name as merchant_url, b.formatted_location, b.service_name, e.first_name as employee_first_name, e.last_name as employee_last_name
	from "Booking" b
	join "BookingParticipant" bp on bp.booking_id = b.id
	join "Customer" c on bp.customer_id = c.id
	join "User" u on c.user_id = u.id
	join "Merchant" m on b.merchant_id = m.id
	left join "Employee" e on b.employee_id = e.id
	where u.id = $1 and b.from_date > now() and b.status in ('booked', 'confirmed') and (b.from_date, b.id) > ($3, $4)
	order by b.from_date asc, b.id asc
	limit $2
	`

	rows, _ := r.db.Query(ctx, query, userId, limit, cursorStart, cursorId)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingForUser])
	if err != nil {
		return []domain.BookingForUser{}, fmt.Errorf("GetUpcomingBookingsForUser: %w", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetCompletedBookingsForUser(ctx context.Context, userId uuid.UUID, limit int, cursorStart time.Time, cursorId int) ([]domain.BookingForUser, error) {
	query := `
	select b.id, b.status, b.booking_type, b.is_recurring, b.from_date, b.to_date, b.price_per_person, m.name as merchant_name,
		m.url_name as merchant_url, b.formatted_location, b.service_name, e.first_name as employee_first_name, e.last_name as employee_last_name
	from "Booking" b
	join "BookingParticipant" bp on bp.booking_id = b.id
	join "Customer" c on bp.customer_id = c.id
	join "User" u on c.user_id = u.id
	join "Merchant" m on b.merchant_id = m.id
	left join "Employee" e on b.employee_id = e.id
	where u.id = $1 and b.to_date < now() and b.status not in ('cancelled', 'no-show') and (b.from_date, b.id) > ($3, $4)
	order by b.from_date desc, b.id desc
	limit $2
	`

	rows, _ := r.db.Query(ctx, query, userId, limit, cursorStart, cursorId)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingForUser])
	if err != nil {
		return []domain.BookingForUser{}, fmt.Errorf("GetCompletedBookingsForUser: %w", err)
	}

	return bookings, nil
}

// TODO: order by updated_at desc
func (r *bookingRepository) GetCancelledBookingsForUser(ctx context.Context, userId uuid.UUID, limit int, cursorStart time.Time, cursorId int) ([]domain.BookingForUser, error) {
	query := `
	select b.id, b.status, b.booking_type, b.is_recurring, b.from_date, b.to_date, b.price_per_person, m.name as merchant_name,
		m.url_name as merchant_url, b.formatted_location, b.service_name, e.first_name as employee_first_name, e.last_name as employee_last_name
	from "Booking" b
	join "BookingParticipant" bp on bp.booking_id = b.id
	join "Customer" c on bp.customer_id = c.id
	join "User" u on c.user_id = u.id
	join "Merchant" m on b.merchant_id = m.id
	left join "Employee" e on b.employee_id = e.id
	where u.id = $1 and b.status in ('cancelled') and b.cancelled_by_merchant_on is not null and (b.from_date, b.id) > ($3, $4)
	order by b.id asc
	limit $2
	`

	rows, _ := r.db.Query(ctx, query, userId, limit, cursorStart, cursorId)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingForUser])
	if err != nil {
		return []domain.BookingForUser{}, fmt.Errorf("GetCancelledBookingsForUser: %w", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetBookingPhases(ctx context.Context, bookingId int) ([]domain.BookingPhase, error) {
	query := `
	select *
	from "BookingPhase"
	where booking_id = $1
	`

	rows, _ := r.db.Query(ctx, query, bookingId)
	bookingPhases, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingPhase])
	if err != nil {
		return []domain.BookingPhase{}, fmt.Errorf("GetBookingPhases: %w", err)
	}

	return bookingPhases, nil
}

func (r *bookingRepository) GetBookingCancelDeadline(ctx context.Context, bookingId int) (int, error) {
	query := `
	select coalesce(s.cancel_deadline, m.cancel_deadline) as cancel_deadline
	from "Booking" b
	join "Merchant" m on m.id = b.merchant_id
	left join "Service" s on s.id = b.service_id
	where b.id = $1
	`

	var cancelDeadline int

	err := r.db.QueryRow(ctx, query, bookingId).Scan(&cancelDeadline)
	if err != nil {
		return 0, fmt.Errorf("GetBookingCancelDeadline: %w", err)
	}

	return cancelDeadline, nil
}

func (r *bookingRepository) GetReservedTimes(ctx context.Context, merchant_id uuid.UUID, location_id int, day time.Time) ([]domain.BookingSlot, error) {
	query := `
    select bp.from_date, bp.to_date
	from "BookingPhase" bp
	join "Booking" b on bp.booking_id = b.id
    where b.merchant_id = $1 and b.location_id = $2 and DATE(b.from_date) = $3 and b.status not in ('cancelled', 'completed') and bp.phase_type = 'active'
    ORDER BY bp.from_date`

	rows, _ := r.db.Query(ctx, query, merchant_id, location_id, day)
	reservedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingSlot])
	if err != nil {
		return nil, fmt.Errorf("GetReservedTimes: %w", err)
	}

	return reservedTimes, nil
}

func (r *bookingRepository) GetReservedTimesForPeriod(ctx context.Context, merchantId uuid.UUID, locationId int, startDate time.Time, endDate time.Time) ([]domain.BookingSlot, error) {
	query := `
	select bp.from_date, bp.to_date
	from "BookingPhase" bp
	join "Booking" b on bp.booking_id = b.id
	where b.merchant_id = $1 and b.location_id = $2 and DATE(b.from_date) >= $3 and DATE(b.to_date) <= $4
		and b.status not in ('cancelled', 'completed') and bp.phase_type = 'active'
	order by bp.from_date`

	rows, _ := r.db.Query(ctx, query, merchantId, locationId, startDate, endDate)
	reservedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingSlot])
	if err != nil {
		return nil, fmt.Errorf("GetReservedTimesForPeriod: %w", err)
	}

	return reservedTimes, nil
}

func (r *bookingRepository) GetAvailableGroupBookingsForPeriod(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int, startTime time.Time, endTime time.Time) ([]domain.BookingSlot, error) {
	query := `
	select b.from_date, b.to_date from "Booking" b
	where b.booking_type in ('event', 'class') and b.merchant_id = $1 and b.service_id = $2 and b.location_id = $3 and DATE(b.from_date) >= $4 and DATE(b.to_date) <= $5
		and b.status not in ('cancelled', 'completed') and b.current_participants < b.max_participants
	order by b.from_date
	`

	rows, _ := r.db.Query(ctx, query, merchantId, serviceId, locationId, startTime, endTime)
	availableBookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingSlot])
	if err != nil {
		return nil, fmt.Errorf("GetAvailableGroupBookingsForPeriod: %w", err)
	}

	return availableBookings, nil
}

func (r *bookingRepository) GetClosestAvailableGroupBooking(ctx context.Context, merchantId uuid.UUID, serviceId, locationId int, searchStart, searchEnd time.Time) (domain.Booking, error) {
	query := `
	select *
	from "Booking"
	where merchant_id = $1 and service_id = $2 and location_id = $3 and from_date >= $4 and to_date <= $5 and current_participants < max_participants and status not in ('cancelled', 'completed')
	order by from_date asc
	limit 1`

	row, _ := r.db.Query(ctx, query, merchantId, serviceId, locationId, searchStart, searchEnd)
	booking, err := pgx.CollectExactlyOneRow(row, pgx.RowToStructByName[domain.Booking])
	if err != nil {
		return domain.Booking{}, fmt.Errorf("GetClosestAvailableGroupBooking: %w", err)
	}

	return booking, nil
}

func (r *bookingRepository) NewBookingSeries(ctx context.Context, bs domain.BookingSeries) (domain.BookingSeries, error) {
	query := `
	insert into "BookingSeries" (booking_type, merchant_id, employee_id, service_id, location_id, rrule, dstart, timezone, is_active, generated_until,
		service_name, price_per_person, total_price, price_type, formatted_location, min_participants, max_participants, current_participants)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	returning *
	`

	rows, _ := r.db.Query(ctx, query, bs.BookingType, bs.MerchantId, bs.EmployeeId, bs.ServiceId, bs.LocationId, bs.Rrule, bs.Dstart, bs.Timezone, true, bs.GeneratedUntil,
		bs.ServiceName, bs.PricePerPerson, bs.TotalPrice, bs.PriceType, bs.FormattedLocation, bs.MinParticipants, bs.MaxParticipants, bs.CurrentParticipants)
	bookingSeries, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.BookingSeries])
	if err != nil {
		return domain.BookingSeries{}, fmt.Errorf("NewBookingSeries: %w", err)
	}

	return bookingSeries, nil
}

func (r *bookingRepository) NewBookingSeriesParticipants(ctx context.Context, bookingSeriesParticipants []domain.BookingSeriesParticipant) ([]domain.BookingSeriesParticipant, error) {
	query := `
	insert into "BookingSeriesParticipant" (booking_series_id, customer_id, is_active)
	select unnest($1::int[]), unnest($2::uuid[]), unnest($3::boolean[])
	returning *
	`

	seriesIds := make([]int, len(bookingSeriesParticipants))
	customerIds := make([]pgtype.UUID, len(bookingSeriesParticipants))
	isActives := make([]bool, len(bookingSeriesParticipants))

	for i, bsp := range bookingSeriesParticipants {
		seriesIds[i] = bsp.BookingSeriesId
		if bsp.CustomerId == nil {
			customerIds[i] = pgtype.UUID{Valid: false}
		} else {
			customerIds[i] = pgtype.UUID{Bytes: *bsp.CustomerId, Valid: true}
		}
		isActives[i] = bsp.IsActive
	}

	rows, _ := r.db.Query(ctx, query, seriesIds, customerIds, isActives)
	bookingSeriesParticipants, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingSeriesParticipant])
	if err != nil {
		return []domain.BookingSeriesParticipant{}, fmt.Errorf("NewBookingSeriesParticipants: %w", err)
	}

	return bookingSeriesParticipants, nil
}

func (r *bookingRepository) NewBookingSeriesPhases(ctx context.Context, bookingSeriesPhases []domain.BookingSeriesPhase) error {
	query := `
	insert into "BookingSeriesPhase" (booking_series_id, service_phase_id, name, sequence, duration, phase_type)
	select unnest($1::int[]), unnest($2::int[]), unnest($3::text[]), unnest($4::int[]), unnest($5::int[]), unnest($6::text[])
	`

	bookingSeriesPhaseCount := len(bookingSeriesPhases)

	bookingSeriesIds := make([]int, bookingSeriesPhaseCount)
	servicePhaseIds := make([]pgtype.Int4, bookingSeriesPhaseCount)
	names := make([]string, bookingSeriesPhaseCount)
	sequences := make([]int, bookingSeriesPhaseCount)
	durations := make([]int, bookingSeriesPhaseCount)
	phaseTypes := make([]string, bookingSeriesPhaseCount)

	for i, bsp := range bookingSeriesPhases {
		bookingSeriesIds[i] = bsp.BookingSeriesId
		if bsp.ServicePhaseId == nil {
			servicePhaseIds[i] = pgtype.Int4{Valid: false}
		} else {
			servicePhaseIds[i] = pgtype.Int4{Int32: int32(*bsp.ServicePhaseId), Valid: true}
		}
		names[i] = bsp.Name
		sequences[i] = bsp.Sequence
		durations[i] = bsp.Duration
		phaseTypes[i] = bsp.PhaseType.String()
	}

	_, err := r.db.Exec(ctx, query, bookingSeriesIds, servicePhaseIds, names, sequences, durations, phaseTypes)
	if err != nil {
		return fmt.Errorf("NewBookingSeriesPhases: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingSeriesRrule(ctx context.Context, seriesId int, rrule string, dstart time.Time) (int, error) {
	query := `
	update "BookingSeries"
	set rrule = $2, dstart = $3, version = version + 1
	where id = $1
	returning version
	`

	var seriesVersion int

	err := r.db.QueryRow(ctx, query, seriesId, rrule, dstart).Scan(&seriesVersion)
	if err != nil {
		return 0, fmt.Errorf("UpdateBookingSeriesRrule: %w", err)
	}

	return seriesVersion, nil
}

func (r *bookingRepository) UpdateBookingSeriesGeneratedUntil(ctx context.Context, seriesId int, generatedUntil time.Time) error {
	query := `
	update "BookingSeries"
	set generated_until = $2
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, seriesId, generatedUntil)
	if err != nil {
		return fmt.Errorf("UpdateBookingSeriesGeneratedUntil: %w", err)
	}

	return nil
}

func (r *bookingRepository) DeactivateBookingSeries(ctx context.Context, seriesId int) error {
	query := `
	update "BookingSeries"
	set is_active = false
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, seriesId)
	if err != nil {
		return fmt.Errorf("DeactivateBookingSeries: %w", err)
	}

	return nil
}

func (r *bookingRepository) UpdateBookingSeriesDetails(ctx context.Context, seriesId int, details domain.BookingDetails) error {
	query := `
	update "BookingSeries"
	set price_per_person = $2, total_price = $3, min_participants = $4, max_participants = $5, current_participants = $6
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, seriesId, details.PricePerPerson, details.TotalPrice,
		details.MinParticipants, details.MaxParticipants, details.CurrentParticipants)
	if err != nil {
		return fmt.Errorf("UpdateBookingSeriesDetails: %w", err)
	}

	return nil
}

func (r *bookingRepository) DeleteBookingSeriesParticipants(ctx context.Context, seriesId int, customerIds []uuid.UUID) error {
	query := `delete from "BookingSeriesParticipant"
	where booking_series_id = $1 and customer_id = any($2::uuid[])`

	_, err := r.db.Exec(ctx, query, seriesId, customerIds)
	if err != nil {
		return fmt.Errorf("DeleteBookingSeriesParticipants: %w", err)
	}

	return nil
}

func (r *bookingRepository) GetFutureSeriesBookingsWithLock(ctx context.Context, seriesId, fromOccurrenceIndex, limit int) ([]domain.Booking, error) {
	query := `
	select * from "Booking"
	where booking_series_id = $1 and occurrence_index > $2
	order by occurrence_index asc
	limit $3
	for update
	`

	rows, _ := r.db.Query(ctx, query, seriesId, fromOccurrenceIndex, limit)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Booking])
	if err != nil {
		return nil, fmt.Errorf("GetFutureSeriesBookingsWithLock: %w", err)
	}

	return bookings, nil
}

func (r *bookingRepository) GetBookingSeries(ctx context.Context, seriesId int) (domain.BookingSeries, error) {
	query := `
	select *
	from "BookingSeries"
	where id = $1
	`

	rows, _ := r.db.Query(ctx, query, seriesId)
	bookingSeries, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingSeries])
	if err != nil {
		return domain.BookingSeries{}, fmt.Errorf("GetBookingSeries: %w", err)
	}
	return bookingSeries, nil
}

func (r *bookingRepository) GetActiveBookingSeriesIds(ctx context.Context, tresholdTime time.Time) ([]int, error) {
	query := `
	select id
	from "BookingSeries"
	where is_active = true and (generated_until < $1 or generated_until is null)
	`

	rows, _ := r.db.Query(ctx, query, tresholdTime)
	ids, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return []int{}, fmt.Errorf("GetActiveBookingSeriesIds: %w", err)
	}

	return ids, nil
}

func (r *bookingRepository) GetSeriesLastOccurrenceIndex(ctx context.Context, seriesId int) (int, error) {
	query := `
	select coalesce(max(occurrence_index), 0)
	from "Booking"
	where booking_series_id = $1
	`

	var occurrenceIndex int

	err := r.db.QueryRow(ctx, query, seriesId).Scan(&occurrenceIndex)
	if err != nil {
		return 0, fmt.Errorf("GetSeriesLastOccurrenceIndex: %w", err)
	}

	return occurrenceIndex, err
}

func (r *bookingRepository) GetSeriesOccurrenceDateByIndex(ctx context.Context, seriesId int, occurrenceIndex int) (time.Time, error) {
	query := `
	select series_original_date
	from "Booking"
	where booking_series_id = $1 and occurrence_index = $2
	`

	var occurrenceDate time.Time

	err := r.db.QueryRow(ctx, query, seriesId, occurrenceIndex).Scan(&occurrenceDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("GetSeriesOccurrenceDateByIndex: %w", err)
	}

	return occurrenceDate, nil
}

func (r *bookingRepository) GetBookingSeriesParticipants(ctx context.Context, seriesId int) ([]domain.BookingSeriesParticipant, error) {
	query := `
	select *
	from "BookingSeriesParticipant"
	where booking_series_id = $1
	`

	rows, _ := r.db.Query(ctx, query, seriesId)
	participants, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingSeriesParticipant])
	if err != nil {
		return []domain.BookingSeriesParticipant{}, fmt.Errorf("GetBookingSeriesParticipants: %w", err)
	}

	return participants, nil
}

func (r *bookingRepository) GetBookingSeriesPhases(ctx context.Context, seriesId int) ([]domain.BookingSeriesPhase, error) {
	query := `
	select *
	from "BookingSeriesPhase"
	where booking_series_id = $1
	`

	rows, _ := r.db.Query(ctx, query, seriesId)
	bookingSeriesPhases, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingSeriesPhase])
	if err != nil {
		return []domain.BookingSeriesPhase{}, fmt.Errorf("GetBookingSeriesPhases: %w", err)
	}

	return bookingSeriesPhases, nil
}
