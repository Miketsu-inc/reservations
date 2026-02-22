package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type bookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) domain.BookingRepository {
	return &bookingRepository{db: db}
}

type newBookingParticipantData struct {
	CustomerId   *uuid.UUID `json:"customer_id"`
	CustomerNote *string    `json:"customer_note"`
}

type newBookingData struct {
	Status          types.BookingStatus         `json:"status"`
	BookingType     types.BookingType           `json:"booking_type"`
	MerchantId      uuid.UUID                   `json:"merchant_id"`
	ServiceId       int                         `json:"service_id"`
	LocationId      int                         `json:"location_id"`
	FromDate        time.Time                   `json:"from_date"`
	ToDate          time.Time                   `json:"to_date"`
	MerchantNote    *string                     `json:"merchant_note"`
	PricePerPerson  currencyx.Price             `json:"price_per_person"`
	CostPerPerson   currencyx.Price             `json:"cost_per_person"`
	MinParticipants int                         `json:"min_participants"`
	MaxParticipants int                         `json:"max_participants"`
	Participants    []newBookingParticipantData `json:"participants"`
	Phases          []domain.PublicServicePhase `json:"phases"`
}

func newBooking(ctx context.Context, tx pgx.Tx, nb newBookingData) (int, error) {
	var bookingId int

	insertBookingQuery := `
		insert into "Booking" (status, booking_type, merchant_id, service_id, location_id, from_date, to_date)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning id
		`

	err := tx.QueryRow(ctx, insertBookingQuery, nb.Status, nb.BookingType, nb.MerchantId, nb.ServiceId, nb.LocationId, nb.FromDate, nb.ToDate).Scan(&bookingId)
	if err != nil {
		return bookingId, err
	}

	insertBookingPhaseQuery := `
		insert into "BookingPhase" (booking_id, service_phase_id, from_date, to_date)
		values ($1, $2, $3, $4)
		`

	bookingStart := nb.FromDate
	for _, phase := range nb.Phases {
		phaseDuration := time.Duration(phase.Duration) * time.Minute
		bookingEnd := bookingStart.Add(phaseDuration)

		_, err = tx.Exec(ctx, insertBookingPhaseQuery, bookingId, phase.Id, bookingStart, bookingEnd)
		if err != nil {
			return 0, err
		}

		bookingStart = bookingEnd
	}

	insertBookingDetailsQuery := `
		insert into "BookingDetails" (booking_id, price_per_person, cost_per_person, total_price, total_cost, merchant_note, min_participants, max_participants, current_participants)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`

	participantCount := len(nb.Participants)
	if participantCount == 0 {
		participantCount = 1
	}

	countStr := strconv.Itoa(participantCount)

	totalPrice, err := nb.PricePerPerson.Mul(countStr)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total price: %w", err)
	}

	totalCost, err := nb.CostPerPerson.Mul(countStr)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	_, err = tx.Exec(ctx, insertBookingDetailsQuery, bookingId, nb.PricePerPerson, nb.CostPerPerson, totalPrice, totalCost, nb.MerchantNote, nb.MinParticipants, nb.MaxParticipants, participantCount)
	if err != nil {
		return 0, err
	}

	if nb.BookingType == types.BookingTypeAppointment {
		if len(nb.Participants) > 1 {
			return 0, fmt.Errorf("appointment type bookings must not have more than 1 participant")
		}
	}

	// needed for unnest to work with nullable types in pgx
	var participantIds []pgtype.UUID
	var participantNotes []pgtype.Text
	for _, bp := range nb.Participants {
		if bp.CustomerId == nil {
			participantIds = append(participantIds, pgtype.UUID{Valid: false})
		} else {
			participantIds = append(participantIds, pgtype.UUID{
				Valid: true,
				Bytes: *bp.CustomerId,
			})
		}

		if bp.CustomerNote == nil {
			participantNotes = append(participantNotes, pgtype.Text{Valid: false})
		} else {
			participantNotes = append(participantNotes, pgtype.Text{
				Valid:  true,
				String: *bp.CustomerNote,
			})
		}
	}

	batchInsertBookingParticipantQuery := `
		insert into "BookingParticipant" (status, booking_id, customer_id, customer_note)
		select $1, $2, unnest($3::uuid[]), unnest($4::text[])
		`

	_, err = tx.Exec(ctx, batchInsertBookingParticipantQuery, nb.Status, bookingId, participantIds, participantNotes)
	if err != nil {
		return 0, err
	}

	return bookingId, nil
}

func (r *bookingRepository) NewBookingByCustomer(ctx context.Context, nb domain.NewCustomerBooking) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	var IsBlacklisted bool
	ensureCustomerQuery := `
	insert into "Customer" (id, merchant_id, user_id) values ($1, $2, $3)
	on conflict (merchant_id, user_id) do update
	set merchant_id = excluded.merchant_id
	returning id, is_blacklisted`

	var customerId uuid.UUID
	err = tx.QueryRow(ctx, ensureCustomerQuery, nb.CustomerId, nb.MerchantId, nb.UserId).Scan(&customerId, &IsBlacklisted)
	if err != nil {
		return 0, err
	}
	if IsBlacklisted {
		return 0, fmt.Errorf("you are blacklisted, please contact the merchant by email or phone to make a booking")
	}

	var bookingId int
	if nb.BookingType == types.BookingTypeAppointment {
		bookingId, err = newBooking(ctx, tx, newBookingData{
			Status:          nb.Status,
			BookingType:     nb.BookingType,
			MerchantId:      nb.MerchantId,
			ServiceId:       nb.ServiceId,
			LocationId:      nb.LocationId,
			FromDate:        nb.FromDate,
			ToDate:          nb.ToDate,
			MerchantNote:    nil,
			PricePerPerson:  nb.PricePerPerson,
			CostPerPerson:   nb.CostPerPerson,
			MinParticipants: 1,
			MaxParticipants: 1,
			Participants: []newBookingParticipantData{{
				CustomerId:   &customerId,
				CustomerNote: nb.CustomerNote,
			}},
			Phases: nb.Phases,
		})
		if err != nil {
			return 0, err
		}
	} else {
		if nb.BookingId == nil {
			return 0, fmt.Errorf("booking_id is required for a class or event")
		}

		bookingId = *nb.BookingId
		addParticipantQuery := `
		update "BookingDetails" bd
		set current_participants = current_participants + 1
		from "Booking" b
		where b.id = bd.booking_id and b.id = $1 and b.booking_type in ('event', 'class') and b.status not in ('cancelled', 'completed') and bd.current_participants < bd.max_participants
		returning bd.total_price, bd.total_cost
		`

		var totalPrice, totalCost currencyx.Price

		err := tx.QueryRow(ctx, addParticipantQuery, bookingId).Scan(&totalPrice, &totalCost)
		if err != nil {
			return 0, err
		}

		insertParticipantQuery := `
		insert into "BookingParticipant" (status, booking_id, customer_id, customer_note)
		values ($1, $2, $3, $4)`

		_, err = tx.Exec(ctx, insertParticipantQuery, nb.Status, bookingId, customerId, nb.CustomerNote)
		if err != nil {
			return 0, err
		}

		newTotalPrice, err := totalPrice.Add(nb.PricePerPerson.Amount)
		if err != nil {
			return 0, err
		}
		newTotalCost, err := totalCost.Add(nb.CostPerPerson.Amount)
		if err != nil {
			return 0, err
		}

		updateDetailsQuery := `
		update "BookingDetails" set total_price = $2, total_cost = $3
		where booking_id = $1`

		_, err = tx.Exec(ctx, updateDetailsQuery, bookingId, newTotalPrice, newTotalCost)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return bookingId, nil
}

func (r *bookingRepository) NewBookingByMerchant(ctx context.Context, nb domain.NewMerchantBooking) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	var participants []newBookingParticipantData
	for _, customerId := range nb.Participants {
		participants = append(participants, newBookingParticipantData{CustomerId: customerId, CustomerNote: nil})
	}

	bookingId, err := newBooking(ctx, tx, newBookingData{
		Status:          nb.Status,
		BookingType:     nb.BookingType,
		MerchantId:      nb.MerchantId,
		ServiceId:       nb.ServiceId,
		LocationId:      nb.LocationId,
		FromDate:        nb.FromDate,
		ToDate:          nb.ToDate,
		MerchantNote:    nb.MerchantNote,
		PricePerPerson:  nb.PricePerPerson,
		CostPerPerson:   nb.CostPerPerson,
		MinParticipants: nb.MinParticipants,
		MaxParticipants: nb.MaxParticipants,
		Participants:    participants,
		Phases:          nb.Phases,
	})
	if err != nil {
		return 0, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return bookingId, nil
}

func (r *bookingRepository) UpdateBookingData(ctx context.Context, merchantId uuid.UUID, bookingId int, merchant_note string, offset time.Duration) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	bookingDetailsQuery := `
	update "BookingDetails" bd
	set merchant_note = $3
	from "Booking" b
	where b.id = $1 and b.id = bd.booking_id and b.merchant_id = $2 and b.status not in ('cancelled', 'completed')
	`
	_, err = tx.Exec(ctx, bookingDetailsQuery, bookingId, merchantId, merchant_note)
	if err != nil {
		return err
	}

	bookingQuery := `
	update "Booking"
	set from_date = from_date + $3, to_date = to_date + $3
	where id = $1 and merchant_id = $2 and status not in ('cancelled', 'completed')
	`

	_, err = tx.Exec(ctx, bookingQuery, bookingId, merchantId, offset)
	if err != nil {
		return err
	}

	bookingPhaseQuery := `
	update "BookingPhase"
	set from_date = from_date + $2, to_date = to_date + $2
	where booking_id = $1
	`

	_, err = tx.Exec(ctx, bookingPhaseQuery, bookingId, offset)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *bookingRepository) GetCalendarEventsByMerchant(ctx context.Context, merchantId uuid.UUID, start string, end string) (domain.CalendarEvents, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	// nolint:errcheck
	defer tx.Rollback(ctx)

	var events domain.CalendarEvents

	bookingQuery := `
	select b.id, b.from_date, b.to_date, bp.customer_note, bd.merchant_note, bd.total_price as price, bd.total_cost as cost,
		s.name as service_name, s.color as service_color, s.total_duration as service_duration,
		coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name, coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	join "Service" s on b.service_id = s.id
	join "BookingDetails" bd on bd.booking_id = b.id
	left join "BookingParticipant" bp on bp.booking_id = b.id
	left join "Customer" c on bp.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.to_date <= $3 AND b.status not in ('cancelled')
	order by b.id
	`
	rows, _ := tx.Query(ctx, bookingQuery, merchantId, start, end)
	events.Bookings, err = pgx.CollectRows(rows, pgx.RowToStructByName[domain.PublicBookingDetails])
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	blockedTimeQuery := `
	select b.id, b.employee_id, b.name, b.from_date, b.to_date, b.all_day, btt.icon, btt.id as blocked_type_id from "BlockedTime" b
	left join "BlockedTimeType" btt on btt.id = b.blocked_type_id
	where b.merchant_id = $1 and b.from_date <= $3 and b.to_date >= $2
	order by b.id
	`

	rows, _ = tx.Query(ctx, blockedTimeQuery, merchantId, start, end)
	events.BlockedTimes, err = pgx.CollectRows(rows, pgx.RowToStructByName[domain.BlockedTimeEvent])
	if err != nil {
		return domain.CalendarEvents{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.CalendarEvents{}, err
	}

	return events, nil
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

// TODO: what should the booking participant status be here?
func (r *bookingRepository) CancelBookingByMerchant(ctx context.Context, merchantId uuid.UUID, bookingId int, cancellationReason string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	bookingDetailsQuery := `
	update "BookingDetails" bd
	set cancelled_by_merchant_on = $1, cancellation_reason = $2
	from "Booking" b
	where b.id = $4 and b.id = bd.booking_id and b.merchant_id = $3 and b.status not in ('cancelled', 'completed') and b.from_date > $1
	`

	_, err = tx.Exec(ctx, bookingDetailsQuery, time.Now().UTC(), cancellationReason, merchantId, bookingId)
	if err != nil {
		return err
	}

	bookingQuery := `
	update "Booking"
	set status = 'cancelled'
	where id = $1 and merchant_id = $2
	`

	_, err = tx.Exec(ctx, bookingQuery, bookingId, merchantId)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *bookingRepository) CancelBookingByCustomer(ctx context.Context, customerId uuid.UUID, bookingId int) (uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	cancellationTime := time.Now().UTC()

	bookingParticipantQuery := `
	update "BookingParticipant" bp
	set cancelled_on = $1, status = ('cancelled')
	from "Booking" b
	join "BookingDetails" bd on b.id = bd.booking_id
	where bp.customer_id = $2 and bp.booking_id = $3 and b.id = $3 and bp.status not in ('cancelled', 'completed') and b.status not in ('cancelled', 'completed') and b.from_date > $1
	returning bp.email_id, bd.price_per_person, bd.cost_per_person, bd.total_price, bd.total_cost
	`

	var emailId *uuid.UUID
	var pricePerPerson, costPerPerson, totalPrice, totalCost currencyx.Price
	err = tx.QueryRow(ctx, bookingParticipantQuery, cancellationTime, customerId, bookingId).Scan(&emailId, &pricePerPerson, &costPerPerson, &totalPrice, &totalCost)
	if err != nil {
		return uuid.Nil, err
	}

	bookingDetailsQuery := `
	update "BookingDetails" bd
	set current_participants = current_participants - 1, total_price = $3, total_cost = $4
	from "Booking" b
	where b.id = bd.booking_id and b.id = $2 and b.status not in ('cancelled', 'completed') and b.from_date > $1
	returning b.booking_type
	`

	newTotalPrice, err := totalPrice.Sub(pricePerPerson.Amount)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to calculate total price: %w", err)
	}
	newTotalCost, err := totalCost.Sub(costPerPerson.Amount)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	var bookingType types.BookingType
	err = tx.QueryRow(ctx, bookingDetailsQuery, cancellationTime, bookingId, newTotalPrice, newTotalCost).Scan(&bookingType)
	if err != nil {
		return uuid.Nil, err
	}

	if bookingType == types.BookingTypeAppointment {
		bookingQuery := `
		update "Booking"
		set status = 'cancelled'
		where id = $1 and booking_type = 'appointment' and status not in ('cancelled', 'completed') and from_date > $2
		`

		_, err = tx.Exec(ctx, bookingQuery, bookingId, cancellationTime)
		if err != nil {
			return uuid.Nil, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	// in case there was no email scheduled (like booked less then 24 hour before the start)
	if emailId == nil {
		return uuid.Nil, nil
	}

	return *emailId, nil
}

func (r *bookingRepository) UpdateEmailIdForBooking(ctx context.Context, bookingId int, emailId string, customerId uuid.UUID) error {
	emailUUID, err := uuid.Parse(emailId)
	if err != nil {
		return err
	}

	query := `
	update "BookingParticipant"
	set email_id = $1
	from "Booking" b
	where booking_id = $2 and customer_id = $3
	`

	_, err = r.db.Exec(ctx, query, emailUUID, bookingId, customerId)
	if err != nil {
		return err
	}

	return nil
}

func (r *bookingRepository) GetBookingDataForEmail(ctx context.Context, bookingId int) (domain.BookingEmailData, error) {
	query := `
	select b.from_date, b.to_date, s.name as service_name, m.name as merchant_name, coalesce(s.cancel_deadline, m.cancel_deadline) as cancel_deadline, l.formatted_location,
	coalesce(
		jsonb_agg(
			jsonb_build_object(
				'customer_id', c.id,
				'email', coalesce(u.email, c.email),
				'email_id', bp.email_id
			)
		) filter (where bp.id is not null),
		'[]'::jsonb
	) as participants
	from "Booking" b
	join "Service" s on s.id = b.service_id
	left join "BookingParticipant" bp on bp.booking_id = b.id
	left join "Customer" c on c.id = bp.customer_id
	left join "User" u on u.id = c.user_id
	join "Merchant" m on m.id = b.merchant_id
	join "Location" l on l.id = b.location_id
	where b.id = $1
	group by b.id, s.id, m.id, l.id`

	var data domain.BookingEmailData
	var participantJson []byte

	err := r.db.QueryRow(ctx, query, bookingId).Scan(&data.FromDate, &data.ToDate, &data.ServiceName,
		&data.MerchantName, &data.CancelDeadline, &data.FormattedLocation, &participantJson)
	if err != nil {
		return domain.BookingEmailData{}, err
	}

	if len(participantJson) > 0 {
		err = json.Unmarshal(participantJson, &data.Participants)
		if err != nil {
			return domain.BookingEmailData{}, err
		}
	} else {
		data.Participants = []domain.ParticipantEmailData{}
	}

	return data, nil
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

func (r *bookingRepository) BatchCreateRecurringBookings(ctx context.Context, nrb domain.NewRecurringBookings) (int, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	assert.True(len(nrb.FromDates) == len(nrb.ToDates), "Length of fromDates and toDates slices should be equal!", len(nrb.FromDates), len(nrb.ToDates), nrb)
	var bookingIds []int

	insertBookingsQuery := `
		insert into "Booking" (status, booking_type, is_recurring, merchant_id, employee_id, service_id, location_id, booking_series_id, series_original_date, from_date, to_date)
		select $1, $2, $3, $4, $5, $6, $7, $8, unnest($9::timestamptz[]), unnest($10::timestamptz[]), unnest($11::timestamptz[])
		returning id
		`

	rows, _ := tx.Query(ctx, insertBookingsQuery, nrb.BookingStatus, nrb.BookingType, true, nrb.MerchantId, nrb.EmployeeId, nrb.ServiceId, nrb.LocationId,
		nrb.BookingSeriesId, nrb.FromDates, nrb.FromDates, nrb.ToDates)
	bookingIds, err = pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return 0, err
	}

	insertBookingPhasesQuery := `
		insert into "BookingPhase" (booking_id, service_phase_id, from_date, to_date)
		select unnest($1::int[]), unnest($2::int[]), unnest($3::timestamptz[]), unnest($4::timestamptz[])
		`

	assert.True(len(nrb.FromDates) == len(bookingIds), "Length of fromDate and bookingIds slices should be equal!", len(nrb.FromDates), len(bookingIds), nrb)

	var phaseIds []int
	var phaseFromDates []time.Time
	var phaseToDates []time.Time

	bookingStart := nrb.FromDates[0]
	for _, phase := range nrb.Phases {
		phaseDuration := time.Duration(phase.Duration) * time.Minute
		bookingEnd := bookingStart.Add(phaseDuration)

		phaseIds = append(phaseIds, phase.Id)
		phaseFromDates = append(phaseFromDates, bookingStart)
		phaseToDates = append(phaseToDates, bookingEnd)

		bookingStart = bookingEnd
	}

	times := len(bookingIds)

	phaseIds = utils.RepeatSlice(phaseIds, times)
	phaseFromDates = utils.RepeatSlice(phaseFromDates, times)
	phaseToDates = utils.RepeatSlice(phaseToDates, times)
	bookingIdsForPhases := utils.RepeatEach(bookingIds, len(nrb.Phases))

	_, err = tx.Exec(ctx, insertBookingPhasesQuery, bookingIdsForPhases, phaseIds, phaseFromDates, phaseToDates)
	if err != nil {
		return 0, err
	}

	insertBookingDetailsQuery := `
		insert into "BookingDetails" (booking_id, price_per_person, cost_per_person, total_price, total_cost, merchant_note, min_participants, max_participants, current_participants)
		select unnest($1::int[]), $2, $3, $4, $5, $6, $7, $8, $9
		`

	_, err = tx.Exec(ctx, insertBookingDetailsQuery, bookingIds, nrb.Details.PricePerPerson, nrb.Details.CostPerPerson, nrb.Details.TotalPrice, nrb.Details.TotalCost,
		"", nrb.Details.MinParticipants, nrb.Details.MaxParticipants, nrb.Details.CurrentParticipants)
	if err != nil {
		return 0, err
	}

	// needed for unnest to work with nullable types in pgx
	var participantIds []pgtype.UUID
	for _, bp := range nrb.Participants {
		if bp.CustomerId == nil {
			participantIds = append(participantIds, pgtype.UUID{Valid: false})
		} else {
			participantIds = append(participantIds, pgtype.UUID{
				Valid: true,
				Bytes: *bp.CustomerId,
			})
		}
	}

	bookingIdsForParticipants := utils.RepeatEach(bookingIds, len(nrb.Participants))
	participantIdsForBookings := utils.RepeatSlice(participantIds, len(bookingIds))

	insertBookingParticipantsQuery := `
		insert into "BookingParticipant" (status, booking_id, customer_id, customer_note)
		select $1, unnest($2::int[]), unnest($3::uuid[]), $4
		`

	_, err = tx.Exec(ctx, insertBookingParticipantsQuery, types.BookingStatusBooked, bookingIdsForParticipants, participantIdsForBookings, nil)
	if err != nil {
		return 0, err
	}

	return bookingIds[0], tx.Commit(ctx)
}

func (r *bookingRepository) GetExistingOccurrenceDates(ctx context.Context, seriesId int, fromDate, toDate time.Time) ([]time.Time, error) {
	query := `
	select series_original_date
	from "Booking"
	where booking_series_id = $1 and series_original_date >= $2 and series_original_date <= $3
	`

	rows, _ := r.db.Query(ctx, query, seriesId, fromDate, toDate)
	dates, err := pgx.CollectRows(rows, pgx.RowTo[time.Time])
	if err != nil {
		return []time.Time{}, nil
	}

	return dates, nil
}

func (r *bookingRepository) NewBookingSeries(ctx context.Context, nbs domain.NewBookingSeries) (domain.CompleteBookingSeries, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.CompleteBookingSeries{}, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	var booking domain.CompleteBookingSeries

	insertBookingSeriesQuery := `
	insert into "BookingSeries" (booking_type, merchant_id, employee_id, service_id, location_id, rrule, dstart, timezone, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	returning *
	`

	rows, _ := tx.Query(ctx, insertBookingSeriesQuery, nbs.BookingType, nbs.MerchantId, nbs.EmployeeId, nbs.ServiceId, nbs.LocationId, nbs.Rrule, nbs.Dstart,
		nbs.Timezone.String(), true)
	bookingSeries, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.BookingSeries])
	if err != nil {
		return domain.CompleteBookingSeries{}, err
	}

	insertBookingSeriesDetailsQuery := `
	insert into "BookingSeriesDetails" (booking_series_id, price_per_person, cost_per_person, total_price, total_cost, min_participants, max_participants, current_participants)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	returning *
	`

	participantCount := len(nbs.Participants)
	if participantCount == 0 {
		participantCount = 1
	}

	countStr := strconv.Itoa(participantCount)

	totalPrice, err := nbs.PricePerPerson.Mul(countStr)
	if err != nil {
		return domain.CompleteBookingSeries{}, fmt.Errorf("failed to calculate total price: %w", err)
	}

	totalCost, err := nbs.CostPerPerson.Mul(countStr)
	if err != nil {
		return domain.CompleteBookingSeries{}, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	rows, _ = tx.Query(ctx, insertBookingSeriesDetailsQuery, bookingSeries.Id, nbs.PricePerPerson, nbs.CostPerPerson,
		totalPrice, totalCost, nbs.MinParticipants, nbs.MaxParticipants, participantCount)
	bookingSeriesDetails, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.BookingSeriesDetails])
	if err != nil {
		return domain.CompleteBookingSeries{}, err
	}

	// needed for unnest to work with nullable types in pgx
	var participantIds []pgtype.UUID
	for _, uuid := range nbs.Participants {
		if uuid == nil {
			participantIds = append(participantIds, pgtype.UUID{Valid: false})
		} else {
			participantIds = append(participantIds, pgtype.UUID{
				Valid: true,
				Bytes: *uuid,
			})
		}
	}

	insertBookingSeriesParticipantsQuery := `
	insert into "BookingSeriesParticipant" (booking_series_id, customer_id, is_active)
	select $1, unnest($2::uuid[]), $3
	returning *
	`

	rows, _ = tx.Query(ctx, insertBookingSeriesParticipantsQuery, bookingSeries.Id, participantIds, true)
	bookingSeriesParticipants, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.BookingSeriesParticipant])
	if err != nil {
		return domain.CompleteBookingSeries{}, err
	}

	booking.BookingSeries = bookingSeries
	booking.Details = bookingSeriesDetails
	booking.Participants = bookingSeriesParticipants

	err = tx.Commit(ctx)
	if err != nil {
		return domain.CompleteBookingSeries{}, err
	}

	return booking, nil
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
