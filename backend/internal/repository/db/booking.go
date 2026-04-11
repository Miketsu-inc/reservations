package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
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
	insert into "Booking" (status, booking_type, merchant_id, service_id, location_id, from_date, to_date)
	values ($1, $2, $3, $4, $5, $6, $7)
	returning id
	`

	var bookingId int

	err := r.db.QueryRow(ctx, query, booking.Status, booking.BookingType, booking.MerchantId, booking.ServiceId, booking.LocationId, booking.FromDate, booking.ToDate).Scan(&bookingId)
	if err != nil {
		return 0, err
	}

	return bookingId, nil
}

func (r *bookingRepository) NewBookings(ctx context.Context, bookings []domain.Booking) ([]int, error) {
	query := `
	insert into "Booking" (status, booking_type, is_recurring, merchant_id, employee_id, service_id, location_id, booking_series_id, series_original_date, from_date, to_date)
	select unnest($1::booking_status[]), unnest($2::booking_type[]), unnest($3::boolean[]), unnest($4::uuid[]), unnest($5::int[]),
		unnest($6::int[]), unnest($7::int[]), unnest($8::int[]), unnest($9::timestamptz[]), unnest($10::timestamptz[]), unnest($11::timestamptz[])
	returning id
	`

	var bookingIds []int

	statues := make([]string, len(bookings))
	types := make([]string, len(bookings))
	isRecurrings := make([]bool, len(bookings))
	merchantIds := make([]uuid.UUID, len(bookings))
	employeeIds := make([]pgtype.Int4, len(bookings))
	serviceIds := make([]int, len(bookings))
	locationIds := make([]int, len(bookings))
	seriesIds := make([]pgtype.Int4, len(bookings))
	fromDates := make([]time.Time, len(bookings))
	toDates := make([]time.Time, len(bookings))

	for i, b := range bookings {
		statues[i] = b.Status.String()
		types[i] = b.BookingType.String()
		isRecurrings[i] = b.IsRecurring
		merchantIds[i] = b.MerchantId
		if b.EmployeeId == nil {
			employeeIds[i] = pgtype.Int4{Valid: false}
		} else {
			employeeIds[i] = pgtype.Int4{Int32: int32(*b.EmployeeId), Valid: true}
		}
		serviceIds[i] = b.ServiceId
		locationIds[i] = b.LocationId
		if b.BookingSeriesId == nil {
			seriesIds[i] = pgtype.Int4{Valid: false}
		} else {
			seriesIds[i] = pgtype.Int4{Int32: int32(*b.BookingSeriesId), Valid: true}
		}
		fromDates[i] = b.FromDate
		toDates[i] = b.ToDate
	}

	rows, _ := r.db.Query(ctx, query, statues, types, isRecurrings, merchantIds, employeeIds, serviceIds, locationIds,
		seriesIds, fromDates, fromDates, toDates)
	bookingIds, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return []int{}, err
	}

	return bookingIds, nil
}

func (r *bookingRepository) NewBookingPhases(ctx context.Context, bookingPhases []domain.BookingPhase) error {
	query := `
	insert into "BookingPhase" (booking_id, service_phase_id, from_date, to_date)
	select unnest($1::int[]), unnest($2::int[]), unnest($3::timestamptz[]), unnest($4::timestamptz[])
	`

	bookingIds := make([]int, len(bookingPhases))
	servicePhaseIds := make([]int, len(bookingPhases))
	fromDates := make([]time.Time, len(bookingPhases))
	toDates := make([]time.Time, len(bookingPhases))

	for i, bp := range bookingPhases {
		bookingIds[i] = bp.BookingId
		servicePhaseIds[i] = bp.ServicePhaseId
		fromDates[i] = bp.FromDate
		toDates[i] = bp.ToDate
	}

	_, err := r.db.Exec(ctx, query, bookingIds, servicePhaseIds, fromDates, toDates)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) NewBookingDetails(ctx context.Context, bookingDetails domain.BookingDetails) error {
	query := `
	insert into "BookingDetails" (booking_id, price_per_person, cost_per_person, total_price, total_cost, merchant_note, min_participants, max_participants, current_participants)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query, bookingDetails.BookingId, bookingDetails.PricePerPerson, bookingDetails.CostPerPerson, bookingDetails.TotalPrice,
		bookingDetails.TotalCost, bookingDetails.MerchantNote, bookingDetails.MinParticipants, bookingDetails.MaxParticipants, bookingDetails.CurrentParticipants)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) NewBookingDetailsBatch(ctx context.Context, bookingDetails []domain.BookingDetails) error {
	query := `
	insert into "BookingDetails" (booking_id, price_per_person, cost_per_person, total_price, total_cost, merchant_note, min_participants, max_participants, current_participants)
	select unnest($1::int[]), unnest($2::price[]), unnest($3::price[]), unnest($4::price[]), unnest($5::price[]), unnest($6::text[]), unnest($7::int[]), unnest($8::int[]), unnest($9::int[])
	`

	bookingIds := make([]int, len(bookingDetails))
	pricePerPersons := make([]currencyx.Price, len(bookingDetails))
	costPerPersons := make([]currencyx.Price, len(bookingDetails))
	totalPrices := make([]currencyx.Price, len(bookingDetails))
	totalCosts := make([]currencyx.Price, len(bookingDetails))
	merchantNotes := make([]pgtype.Text, len(bookingDetails))
	minParicipants := make([]int, len(bookingDetails))
	maxParicipants := make([]int, len(bookingDetails))
	currentParicipants := make([]int, len(bookingDetails))

	for i, bd := range bookingDetails {
		bookingIds[i] = bd.BookingId
		pricePerPersons[i] = bd.PricePerPerson
		costPerPersons[i] = bd.CostPerPerson
		totalPrices[i] = bd.TotalPrice
		totalCosts[i] = bd.TotalCost
		if bd.MerchantNote == nil {
			merchantNotes[i] = pgtype.Text{Valid: false}
		} else {
			merchantNotes[i] = pgtype.Text{String: *bd.MerchantNote, Valid: true}
		}
		minParicipants[i] = bd.MinParticipants
		maxParicipants[i] = bd.MaxParticipants
		currentParicipants[i] = bd.CurrentParticipants
	}

	_, err := r.db.Exec(ctx, query, bookingIds, pricePerPersons, costPerPersons, totalPrices, totalCosts, merchantNotes, minParicipants,
		maxParicipants, currentParicipants)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) NewBookingParticipants(ctx context.Context, bookingParticipants []domain.BookingParticipant) error {
	query := `
	insert into "BookingParticipant" (status, booking_id, customer_id, customer_note)
	select unnest($1::booking_status[]), unnest($2::int[]), unnest($3::uuid[]), unnest($4::text[])
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
		return err
	}

	return nil
}

func (r *bookingRepository) DeleteBookingPhasesBatch(ctx context.Context, bookingIds []int) error {
	query := `delete from "BookingPhase" where booking_id = any($1::int[])`

	_, err := r.db.Exec(ctx, query, bookingIds)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) DeleteBookingParticipantsBatch(ctx context.Context, bookingIds []int, participantIds []uuid.UUID) error {
	query := `delete from "BookingParticipant"
	where booking_id = any($1::int[]) and customer_id = any($2::uuid[])`

	_, err := r.db.Exec(ctx, query, bookingIds, participantIds)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func (r *bookingRepository) UpdateBookingCoreBatch(ctx context.Context, merchantId uuid.UUID, bookingIds []int, serviceId int, fromDates []time.Time, toDates []time.Time, bookingType types.BookingType, status types.BookingStatus) error {
	query := `
	update "Booking" as b
	set from_date = data.new_from_dates,
	    to_date = data.new_to_dates,
	    service_id = $2,
	    booking_type = $3,
	    status = $4
	from (select unnest($1::int[]) as id, unnest($5::timestamptz[]) as new_from_dates, unnest($6::timestamptz[]) as new_to_dates) as data
	where b.id = data.id and b.merchant_id = $7 and b.status not in ('cancelled', 'completed')
	`

	_, err := r.db.Exec(ctx, query, bookingIds, serviceId, bookingType, status, fromDates, toDates, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) UpdateBookingTotalPrice(ctx context.Context, bookingId int, price, cost currencyx.Price) error {
	query := `
	update "BookingDetails" set total_price = $2, total_cost = $3
	where booking_id = $1
	`

	_, err := r.db.Exec(ctx, query, bookingId, price, cost)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) UpdateBookingDetailsBatch(ctx context.Context, merchantId uuid.UUID, bookingIds []int, details domain.BookingDetails) error {
	query := `
	update "BookingDetails" bd
	set price_per_person = $3, cost_per_person = $4, total_price = $5, total_cost = $6, merchant_note = $7, min_participants = $8, max_participants = $9, current_participants = $10
	from "Booking" b
	where b.id = any($1::int[]) and b.id = bd.booking_id and b.merchant_id = $2 and b.status not in ('cancelled', 'completed')
	`
	_, err := r.db.Exec(ctx, query, bookingIds, merchantId, details.PricePerPerson, details.CostPerPerson, details.TotalPrice,
		details.TotalCost, details.MerchantNote, details.MinParticipants, details.MaxParticipants, details.CurrentParticipants)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) UpdateBookingParticipants(ctx context.Context, participants []domain.BookingParticipant) error {
	query := `
	insert into "BookingParticipant" (booking_id, customer_id, status)
	select unnest($1::int[]), unnest($2::uuid[]), unnest($3::booking_status[])
	on conflict (booking_id, customer_id)
	do update
	set status = excluded.status, cancelled_on = NULL, cancellation_reason = NULL`

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

	_, err := r.db.Exec(ctx, query, bookingIds, customerIds, statuses)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) UpdateParticipantStatus(ctx context.Context, bookingId int, participantId int, status types.BookingStatus) error {
	query := `update "BookingParticipant" set status = $3
	where booking_id = $1 and id = $2`

	_, err := r.db.Exec(ctx, query, bookingId, participantId, status)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) IncrementParticipantCount(ctx context.Context, bookingId int) (currencyx.Price, currencyx.Price, error) {
	query := `
	update "BookingDetails" bd
	set current_participants = current_participants + 1
	from "Booking" b
	where b.id = bd.booking_id and b.id = $1 and b.booking_type in ('event', 'class') and b.status not in ('cancelled', 'completed') and bd.current_participants < bd.max_participants
	returning bd.total_price, bd.total_cost
	`

	var totalPrice, totalCost currencyx.Price

	err := r.db.QueryRow(ctx, query, bookingId).Scan(&totalPrice, &totalCost)
	if err != nil {
		return currencyx.Price{}, currencyx.Price{}, err
	}

	return totalPrice, totalPrice, nil
}

func (r *bookingRepository) DecrementParticipantCount(ctx context.Context, bookingId int) error {
	query := `
	update "BookingDetails" bd
	set current_participants = current_participants - 1
	from "Booking" b
	where b.id = bd.booking_id and b.id = $1 and b.status not in ('cancelled', 'completed') and b.from_date > $2
	`

	var bookingType types.BookingType

	err := r.db.QueryRow(ctx, query, bookingId, time.Now().UTC()).Scan(&bookingType)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) DecrementEveryParticipantCountForCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	update "BookingDetails" bd
	set current_participants = current_participants - 1
	from "Booking" b
	left join "BookingParticipant" bp on b.id = bp.booking_id and bp.customer_id = $1
	where b.id = bd.booking_id and b.merchant_id = $2 and b.booking_type in ('event', 'class')
	`

	_, err := r.db.Exec(ctx, query, customerId, merchantId)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func (r *bookingRepository) CancelBookingByMerchant(ctx context.Context, merchantId uuid.UUID, bookingId int, cancellationReason string) error {
	bookingDetailsQuery := `
	update "BookingDetails" bd
	set cancelled_by_merchant_on = $1, cancellation_reason = $2
	from "Booking" b
	where b.id = $4 and b.id = bd.booking_id and b.merchant_id = $3 and b.status not in ('cancelled', 'completed') and b.from_date > $1
	`

	_, err := r.db.Exec(ctx, bookingDetailsQuery, time.Now().UTC(), cancellationReason, merchantId, bookingId)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) CancelBookingByCustomer(ctx context.Context, bookingId int, customerId uuid.UUID) (types.BookingType, error) {
	query := `
	update "BookingParticipant" bp
	set cancelled_on = $1, status = 'cancelled'
	from "Booking" b
	where bp.customer_id = $2 and bp.booking_id = $1 and b.id = $1 and bp.status not in ('cancelled', 'completed') and b.status not in ('cancelled', 'completed') and b.from_date > $3
	returning b.booking_type
	`

	var bookingType types.BookingType

	err := r.db.QueryRow(ctx, query, bookingId, customerId, time.Now().UTC()).Scan(&bookingType)
	if err != nil {
		return types.BookingType{}, err
	}

	return bookingType, nil
}

func (r *bookingRepository) DeleteAppointmentsByCustomer(ctx context.Context, customerId uuid.UUID, merchantId uuid.UUID) error {
	query := `
	delete from "Booking" b
	using "BookingParticipant" bp
	where bp.booking_id = b.id and bp.customer_id = $1 and b.merchant_id = $2 and b.booking_type = 'appointment'
	`

	_, err := r.db.Exec(ctx, query, customerId, merchantId)
	if err != nil {
		return err
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
		return err
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
		return domain.Booking{}, err
	}

	return booking, nil
}

func (r *bookingRepository) GetPublicBooking(ctx context.Context, bookingId int) (domain.PublicBooking, error) {
	query := `
	select b.from_date, b.to_date, bd.price_per_person as price, m.name as merchant_name, s.name as service_name, m.cancel_deadline, s.price_type,
		b.status = 'cancelled' as is_cancelled,
		l.formatted_location
	from "Booking" b
	join "BookingDetails" bd on bd.booking_id = b.id
	join "Service" s on s.id = b.service_id
	join "Merchant" m on m.id = b.merchant_id
	join "Location" l on l.id = b.location_id
	where b.id = $1
	`

	var data domain.PublicBooking
	err := r.db.QueryRow(ctx, query, bookingId).Scan(&data.FromDate, &data.ToDate, &data.Price, &data.MerchantName,
		&data.ServiceName, &data.CancelDeadline, &data.PriceType, &data.IsCancelled, &data.FormattedLocation)
	if err != nil {
		return domain.PublicBooking{}, err
	}

	return data, nil
}

func (r *bookingRepository) GetLatestBookings(ctx context.Context, merchantId uuid.UUID, afterDate time.Time, rowLimit int) ([]domain.PublicBookingDetails, error) {
	query := `
	select b.id, b.from_date, b.to_date, bp.customer_note, bd.merchant_note, bd.total_price as price, bd.total_cost as cost, s.name as service_name,
		s.color as service_color, s.total_duration as service_duration,
		coalesce(c.first_name, u.first_name) as first_name,
		coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	join "Service" s on b.service_id = s.id
	join "BookingDetails" bd on bd.booking_id = b.id
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
		return []domain.PublicBookingDetails{}, err
	}

	return bookings, nil
}

func (r *bookingRepository) GetUpcomingBookings(ctx context.Context, merchantId uuid.UUID, afterDate time.Time, rowLimit int) ([]domain.PublicBookingDetails, error) {
	query := `
	select b.id, b.from_date, b.to_date, bp.customer_note, bd.merchant_note, bd.total_price as price, bd.total_cost as cost, s.name as service_name,
		s.color as service_color, s.total_duration as service_duration,
		coalesce(c.first_name, u.first_name) as first_name,
		coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	join "Service" s on b.service_id = s.id
	join "BookingDetails" bd on bd.booking_id = b.id
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
		return []domain.PublicBookingDetails{}, err
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
	select b.id, b.booking_type, b.status as booking_status, b.is_recurring, b.from_date, b.to_date, bd.merchant_note, bd.total_price as price, bd.total_cost as cost,
		s.id as service_id, s.name as service_name, s.color as service_color, bd.max_participants,
		coalesce(p.participants, '[]'::jsonb) as participants
	from "Booking" b
	join "Service" s on b.service_id = s.id
	join "BookingDetails" bd on bd.booking_id = b.id
	left join participants p on p.booking_id = b.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.to_date <= $3 AND b.status not in ('cancelled')
	order by b.id
	`

	rows, _ := r.db.Query(ctx, query, merchantId, startTime, endTime)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingForCalendar])
	if err != nil {
		return []domain.BookingForCalendar{}, err
	}

	return bookings, nil
}

func (r *bookingRepository) GetBookingForExternalCalendar(ctx context.Context, bookingId int) (domain.BookingForExternalCalendar, error) {
	query := `
	select b.id, b.status, b.booking_type, b.employee_id, s.name as service_name, s.description as service_description, s.price_type,
		l.formatted_location, b.from_date, b.to_date, bd.total_price, bd.total_cost, bd.merchant_note, bd.current_participants
	from "Booking" b
	join "BookingDetails" bd on b.id = bd.booking_id
	join "Service" s on b.service_id = s.id
	join "Location" l on b.location_id = l.id
	where b.id = $1
	`

	rows, _ := r.db.Query(ctx, query, bookingId)
	bookingData, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingForExternalCalendar])
	if err != nil {
		return domain.BookingForExternalCalendar{}, nil
	}

	return bookingData, nil
}

func (r *bookingRepository) GetBookingForEmail(ctx context.Context, bookingId int, customerId uuid.UUID) (domain.BookingForEmail, error) {
	query := `
	select b.id, b.status, b.from_date, b.to_date, s.name as service_name, s.id as service_id, m.name as merchant_name, m.url_name as merchant_url, m.timezone,
		coalesce(s.cancel_deadline, m.cancel_deadline) as cancel_deadline, l.formatted_location, c.id as customer_id, coalesce(u.email, c.email) as customer_email,
		bp.status as participant_status
	from "BookingParticipant" bp
	join "Booking" b on b.id = bp.booking_id and b.id = $1
	join "Service" s on s.id = b.service_id
	join "Merchant" m on m.id = b.merchant_id
	join "Location" l on l.id = b.location_id
	left join "Customer" c on c.id = bp.customer_id
	left join "User" u on u.id = c.user_id
	where bp.customer_id = $2
	`

	rows, _ := r.db.Query(ctx, query, bookingId, customerId)
	booking, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingForEmail])
	if err != nil {
		return domain.BookingForEmail{}, err
	}

	return booking, nil
}

func (r *bookingRepository) GetBookingDetails(ctx context.Context, bookingId int) (domain.BookingDetails, error) {
	query := `
	select * from "BookingDetails"
	where booking_id = $1
	`

	rows, _ := r.db.Query(ctx, query, bookingId)
	bookingDetails, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingDetails])
	if err != nil {
		return domain.BookingDetails{}, err
	}

	return bookingDetails, nil
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
		return []domain.BookingParticipant{}, err
	}

	return participants, nil
}

func (r *bookingRepository) GetReservedTimes(ctx context.Context, merchant_id uuid.UUID, location_id int, day time.Time) ([]domain.BookingTime, error) {
	query := `
    select bphase.from_date, bphase.to_date
	from "BookingPhase" bphase
	join "Booking" b on bphase.booking_id = b.id
	join "ServicePhase" sp on bphase.service_phase_id = sp.id
    where b.merchant_id = $1 and b.location_id = $2 and DATE(b.from_date) = $3 and b.status not in ('cancelled', 'completed') and sp.phase_type = 'active'
    ORDER BY bphase.from_date`

	rows, _ := r.db.Query(ctx, query, merchant_id, location_id, day)
	reservedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingTime])
	if err != nil {
		return nil, err
	}

	return reservedTimes, nil
}

func (r *bookingRepository) GetReservedTimesForPeriod(ctx context.Context, merchantId uuid.UUID, locationId int, startDate time.Time, endDate time.Time) ([]domain.BookingTime, error) {
	query := `
	select bphase.from_date, bphase.to_date
	from "BookingPhase" bphase
	join "Booking" b on bphase.booking_id = b.id
	join "ServicePhase" sp on bphase.service_phase_id = sp.id
	where b.merchant_id = $1 and b.location_id = $2 and DATE(b.from_date) >= $3 and DATE(b.to_date) <= $4
		and b.status not in ('cancelled', 'completed') and sp.phase_type = 'active'
	order by bphase.from_date`

	rows, _ := r.db.Query(ctx, query, merchantId, locationId, startDate, endDate)
	reservedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingTime])
	if err != nil {
		return nil, err
	}

	return reservedTimes, nil
}

func (r *bookingRepository) GetAvailableGroupBookingsForPeriod(ctx context.Context, merchantId uuid.UUID, serviceId int, locationId int, startTime time.Time, endTime time.Time) ([]domain.BookingTime, error) {
	query := `
	select b.from_date, b.to_date from "Booking" b
	join "BookingDetails" bd on bd.booking_id = b.id
	where b.booking_type in ('event', 'class') and b.merchant_id = $1 and b.service_id = $2 and b.location_id = $3 and DATE(b.from_date) >= $4 and DATE(b.to_date) <= $5
		and b.status not in ('cancelled', 'completed') and bd.current_participants < bd.max_participants
	order by b.from_date
	`

	rows, _ := r.db.Query(ctx, query, merchantId, serviceId, locationId, startTime, endTime)
	availableBookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingTime])
	if err != nil {
		return nil, err
	}

	return availableBookings, nil
}

func (r *bookingRepository) NewBookingSeries(ctx context.Context, bs domain.BookingSeries) (domain.BookingSeries, error) {
	query := `
	insert into "BookingSeries" (booking_type, merchant_id, employee_id, service_id, location_id, rrule, dstart, timezone, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	returning *
	`

	rows, _ := r.db.Query(ctx, query, bs.BookingType, bs.MerchantId, bs.EmployeeId, bs.ServiceId, bs.LocationId, bs.Rrule, bs.Dstart, bs.Timezone, true)
	bookingSeries, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.BookingSeries])
	if err != nil {
		return domain.BookingSeries{}, err
	}

	return bookingSeries, nil
}

func (r *bookingRepository) NewBookingSeriesDetails(ctx context.Context, bsd domain.BookingSeriesDetails) (domain.BookingSeriesDetails, error) {
	query := `
	insert into "BookingSeriesDetails" (booking_series_id, price_per_person, cost_per_person, total_price, total_cost, min_participants, max_participants, current_participants)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	returning *
	`

	rows, _ := r.db.Query(ctx, query, bsd.BookingSeriesId, bsd.PricePerPerson, bsd.CostPerPerson, bsd.TotalPrice, bsd.TotalCost,
		bsd.MinParticipants, bsd.MaxParticipants, bsd.CurrentParticipants)
	bookingSeriesDetails, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.BookingSeriesDetails])
	if err != nil {
		return domain.BookingSeriesDetails{}, err
	}

	return bookingSeriesDetails, nil
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
		return []domain.BookingSeriesParticipant{}, err
	}

	return bookingSeriesParticipants, nil
}

func (r *bookingRepository) UpdateBookingSeriesCore(ctx context.Context, seriesId int, serviceId int, bookingType types.BookingType, rrule string, dstart time.Time) error {
	query := `update "BookingSeries" set service_id = $2, booking_type = $3, rrule = $4, dstart = $5
	where id = $1`

	_, err := r.db.Exec(ctx, query, seriesId, serviceId, bookingType, rrule, dstart)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) UpdateBookingSeriesGeneratedUntil(ctx context.Context, seriesId int, generatedUntil time.Time) error {
	query := `
	update "BookingSeries"
	set generated_until = $2
	where id = $1
	`

	_, err := r.db.Exec(ctx, query, seriesId, generatedUntil)
	if err != nil {
		return err
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
		return err
	}

	return nil
}

func (r *bookingRepository) UpdateBookingSeriesDetails(ctx context.Context, seriesId int, details domain.BookingSeriesDetails) error {
	query := `update "BookingSeriesDetails"
	set price_per_person = $2, cost_per_person = $3, total_price = $4, total_cost = $5, min_participants = $6, max_participants = $7, current_participants = $8
	where booking_series_id = $1`

	_, err := r.db.Exec(ctx, query, seriesId, details.PricePerPerson, details.CostPerPerson, details.TotalPrice, details.TotalCost,
		details.MinParticipants, details.MaxParticipants, details.CurrentParticipants)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) DeleteBookingSeriesParticipants(ctx context.Context, seriesId int, customerIds []uuid.UUID) error {
	query := `delete from "BookingSeriesParticipant"
	where booking_series_id = $1 and customer_id = any($2::uuid[])`

	_, err := r.db.Exec(ctx, query, seriesId, customerIds)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) GetFutureSeriesBookings(ctx context.Context, seriesId int, fromDate time.Time) ([]domain.Booking, error) {
	query := `select * from "Booking"
	where booking_series_id = $1 and from_date >= $2 and status not in ('cancelled')
	order by from_date asc`

	rows, _ := r.db.Query(ctx, query, seriesId, fromDate)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Booking])
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *bookingRepository) GetBookingSeries(ctx context.Context, seriesId int) (domain.BookingSeries, error) {
	query := `select * from "BookingSeries" where id = $1`

	rows, _ := r.db.Query(ctx, query, seriesId)
	bookingSeries, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingSeries])
	if err != nil {
		return domain.BookingSeries{}, err
	}
	return bookingSeries, nil
}

func (r *bookingRepository) GetActiveBookingSeriesIds(ctx context.Context, tresholdTime time.Time) ([]int, error) {
	query := `
	select *
	from "BookingSeries"
	where is_active = true and (generated_until < $1 or generated_until is null)
	`

	rows, _ := r.db.Query(ctx, query, tresholdTime)
	ids, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return []int{}, err
	}

	return ids, nil
}

func (r *bookingRepository) GetBookingSeriesDetails(ctx context.Context, seriesId int) (domain.BookingSeriesDetails, error) {
	query := `
	select *
	from "BookingSeriesDetails"
	where booking_series_id = $1
	`

	rows, _ := r.db.Query(ctx, query, seriesId)
	seriesDetails, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[domain.BookingSeriesDetails])
	if err != nil {
		return domain.BookingSeriesDetails{}, err
	}

	return seriesDetails, nil
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
		return []domain.BookingSeriesParticipant{}, err
	}

	return participants, nil
}
