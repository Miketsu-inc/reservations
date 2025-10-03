package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/booking"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type Booking struct {
	Id          int            `json:"ID" db:"id"`
	Status      booking.Status `json:"status"`
	BookingType booking.Type   `json:"booking_type"`
	MerchantId  uuid.UUID      `json:"merchant_id" db:"merchant_id"`
	ServiceId   int            `json:"service_id" db:"service_id"`
	LocationId  int            `json:"location_id" db:"location_id"`
	FromDate    time.Time      `json:"from_date" db:"from_date"`
	ToDate      time.Time      `json:"to_date" db:"to_date"`
}

type BookingDetails struct {
	Id                    int             `json:"id"`
	BookingId             int             `json:"booking_id"`
	PricePerPerson        currencyx.Price `json:"price_per_person"`
	CostPerPerson         currencyx.Price `json:"cost_per_person"`
	TotalPrice            currencyx.Price `json:"total_price"`
	TotalCost             currencyx.Price `json:"total_cost"`
	MerchantNote          *string         `json:"merchant_note"`
	MinParticipants       int             `json:"min_participants"`
	MaxParticipants       int             `json:"max_participants"`
	CurrentParticipants   int             `json:"current_participants"`
	EmailId               *uuid.UUID      `json:"email_id"`
	CancelledByMerchantOn *time.Time      `json:"cancelled_by_merchant_on"`
	CancellationReason    *string         `json:"cancellation_reason"`
}

type BookingParticipant struct {
	Id                 int            `json:"id"`
	Status             booking.Status `json:"status"`
	BookingId          int            `json:"booking_id"`
	CustomerId         uuid.UUID      `json:"customer_id"`
	CustomerNote       *string        `json:"customer_note"`
	CancelledOn        *time.Time     `json:"cancelled_on"`
	CancellationReason *string        `json:"cancellation_reason"`
	TransferredTo      *uuid.UUID     `json:"transferred_to"`
}

type BookingPhase struct {
	Id             int       `json:"id"`
	BookingId      int       `json:"booking_id"`
	ServicePhaseId int       `json:"service_phase_id"`
	FromDate       time.Time `json:"from_date"`
	ToDate         time.Time `json:"to_date"`
}

type NewBooking struct {
	Status         booking.Status       `json:"status"`
	BookingType    booking.Type         `json:"booking_type"`
	MerchantId     uuid.UUID            `json:"merchant_id"`
	ServiceId      int                  `json:"service_id"`
	LocationId     int                  `json:"location_id"`
	FromDate       time.Time            `json:"from_date"`
	ToDate         time.Time            `json:"to_date"`
	CustomerNote   *string              `json:"customer_note"`
	PricePerPerson currencyx.Price      `json:"price_per_person"`
	CostPerPerson  currencyx.Price      `json:"cost_per_person"`
	UserId         uuid.UUID            `json:"user_id"`
	CustomerId     uuid.UUID            `json:"customer_id"`
	Phases         []PublicServicePhase `json:"phases"`
}

func (s *service) NewBooking(ctx context.Context, nb NewBooking) (int, error) {
	tx, err := s.db.Begin(ctx)
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

	if nb.BookingType == booking.Appointment {
		insertBookingQuery := `
		insert into "Booking" (status, booking_type, merchant_id, service_id, location_id, from_date, to_date)
		values ($1, $2, $3, $4, $5, $6, $7)
		returning id
		`

		err = tx.QueryRow(ctx, insertBookingQuery, nb.Status, nb.BookingType, nb.MerchantId, nb.ServiceId, nb.LocationId, nb.FromDate, nb.ToDate).Scan(&bookingId)
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

		_, err = tx.Exec(ctx, insertBookingDetailsQuery, bookingId, nb.PricePerPerson, nb.CostPerPerson, nb.PricePerPerson, nb.CostPerPerson, "", 1, 1, 1)
		if err != nil {
			return 0, err
		}

		insertBookingParticipantQuery := `
		insert into "BookingParticipant" (status, booking_id, customer_id, customer_note)
		values ($1, $2, $3, $4)
		`

		_, err = tx.Exec(ctx, insertBookingParticipantQuery, nb.Status, bookingId, customerId, nb.CustomerNote)
		if err != nil {
			return 0, err
		}

	} else {
		assert.Never("TODO: Booking events or classes are not implemented yet!", nb)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return 0, err
	}

	return bookingId, nil
}

func (s *service) UpdateBookingData(ctx context.Context, merchantId uuid.UUID, bookingId int, merchant_note string, offset time.Duration) error {
	tx, err := s.db.Begin(ctx)
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
	from_date = from_date + $3, to_date = to_date + $3
	where id = $1 and merchant_id = $2 and b.status not in ('cancelled', 'completed')
	`

	_, err = s.db.Exec(ctx, bookingQuery, bookingId, merchantId, offset)
	if err != nil {
		return err
	}

	bookingPhaseQuery := `
	update "BookingPhase"
	from_date = from_date + $2, to_date = to_date + $2
	where booking_id = $1
	`

	_, err = s.db.Exec(ctx, bookingPhaseQuery, bookingId, offset)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type PublicBookingDetails struct {
	ID              int                      `json:"id" db:"id"`
	FromDate        time.Time                `json:"from_date" db:"from_date"`
	ToDate          time.Time                `json:"to_date" db:"to_date"`
	CustomerNote    string                   `json:"customer_note" db:"customer_note"`
	MerchantNote    string                   `json:"merchant_note" db:"merchant_note"`
	ServiceName     string                   `json:"service_name" db:"service_name"`
	ServiceColor    string                   `json:"service_color" db:"service_color"`
	ServiceDuration int                      `json:"service_duration" db:"service_duration"`
	Price           currencyx.FormattedPrice `json:"price" db:"price"`
	Cost            currencyx.FormattedPrice `json:"cost" db:"cost"`
	FirstName       string                   `json:"first_name" db:"first_name"`
	LastName        string                   `json:"last_name" db:"last_name"`
	PhoneNumber     string                   `json:"phone_number" db:"phone_number"`
}

func (s *service) GetBookingsByMerchant(ctx context.Context, merchantId uuid.UUID, start string, end string) ([]PublicBookingDetails, error) {
	query := `
	select b.id, b.from_date, b.to_date, bp.customer_note, bd.merchant_note, bd.total_price as price, bd.total_cost as cost,
		s.name as service_name, s.color as service_color, s.total_duration as service_duration,
		coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name, coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	join "Service" s on b.service_id = s.id
	join "BookingParticipant" bp on bp.booking_id = b.id
	join "BookingDetails" bd on bd.booking_id = b.id
	join "Customer" c on bp.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.to_date <= $3 AND b.status not in ('cancelled')
	order by b.id
	`

	rows, _ := s.db.Query(ctx, query, merchantId, start, end)
	bookings, err := pgx.CollectRows(rows, pgx.RowToStructByName[PublicBookingDetails])
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

type BookingTime struct {
	From_date time.Time
	To_date   time.Time
}

func (s *service) GetReservedTimes(ctx context.Context, merchant_id uuid.UUID, location_id int, day time.Time) ([]BookingTime, error) {
	query := `
    select bphase.from_date, bphase.to_date
	from "BookingPhase" bphase
	join "Booking" b on bphase.booking_id = b.id
	join "ServicePhase" sp on bphase.service_phase_id = sp.id
    where b.merchant_id = $1 and b.location_id = $2 and DATE(b.from_date) = $3 and b.status not in ('cancelled', 'completed') and sp.phase_type = 'active'
    ORDER BY bphase.from_date`

	rows, _ := s.db.Query(ctx, query, merchant_id, location_id, day)
	reservedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[BookingTime])
	if err != nil {
		return nil, err
	}

	return reservedTimes, nil
}

func (s *service) GetReservedTimesForPeriod(ctx context.Context, merchantId uuid.UUID, locationId int, startDate time.Time, endDate time.Time) ([]BookingTime, error) {
	query := `
	select bphase.from_date, bphase.to_date
	from "BookingPhase" bphase
	join "Booking" b on bphase.booking_id = b.id
	join "ServicePhase" sp on bphase.service_phase_id = sp.id
	where b.merchant_id = $1 and b.location_id = $2 and DATE(b.from_date) >= $3 and DATE(b.to_date) <= $4
		and b.status not in ('cancelled', 'completed') and sp.phase_type = 'active'
	order by bphase.from_date`

	rows, _ := s.db.Query(ctx, query, merchantId, locationId, startDate, endDate)
	reservedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[BookingTime])
	if err != nil {
		return nil, err
	}

	return reservedTimes, nil
}

func (s *service) TransferDummyBookings(ctx context.Context, merchantId uuid.UUID, fromCustomer uuid.UUID, toCustomer uuid.UUID) error {
	query := `
	update "BookingParticipant" bp
	set transferred_to = $3
	from "Booking" b
	join "Customer" c on bp.customer_id = c.id
	where b.merchant_id = $1 and bp.booking_id = b.id and bp.customer_id = $2 and c.user_id is null
	`

	_, err := s.db.Exec(ctx, query, merchantId, fromCustomer, toCustomer)
	if err != nil {
		return err
	}

	return nil
}

// TODO: what should the booking participant status be here?
func (s *service) CancelBookingByMerchant(ctx context.Context, merchantId uuid.UUID, bookingId int, cancellationReason string) error {
	tx, err := s.db.Begin(ctx)
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

func (s *service) CancelBookingByUser(ctx context.Context, customerId uuid.UUID, bookingId int) (uuid.UUID, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	cancellationTime := time.Now().UTC()

	bookingParticipantQuery := `
	update "BookingParticipant" bp
	set cancelled_on = $1
	from "Booking" b
	where bp.customer_id = $2 and bp.booking_id = $3 and bp.status not in ('cancelled', 'completed') and b.status not in ('cancelled', 'completed') and b.from_date > $1
	`

	_, err = tx.Exec(ctx, bookingParticipantQuery, cancellationTime, customerId, bookingId)
	if err != nil {
		return uuid.Nil, err
	}

	bookingDetailsQuery := `
	update "BookingDetails" bd
	set current_participants = current_participants - 1
	from "Booking" b
	where b.id = bd.booking_id and b.id = $2 and b.status not in ('cancelled', 'completed') and b.from_date > $1
	returning bd.email_id
	`

	var emailId uuid.UUID
	err = tx.QueryRow(ctx, bookingDetailsQuery, cancellationTime, bookingId).Scan(&emailId)
	if err != nil {
		return uuid.Nil, err
	}

	bookingQuery := `
	update "Booking"
	set status = 'cancelled'
	where id = $1 and booking_type = 'appointment' and status not in ('cancelled', 'completed') and from_date > $2
	`

	_, err = tx.Exec(ctx, bookingQuery, bookingId, cancellationTime)
	if err != nil {
		return uuid.Nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	return emailId, nil
}

func (s *service) UpdateEmailIdForBooking(ctx context.Context, bookingId int, emailId string) error {
	emailUUID, err := uuid.Parse(emailId)
	if err != nil {
		return err
	}

	query := `
	update "BookingDetails" bd
	set email_id = $1
	from "Booking" b
	where b.id = $2 and bd.booking_id = b.id
	`

	_, err = s.db.Exec(ctx, query, emailUUID, bookingId)
	if err != nil {
		return err
	}

	return nil
}

type BookingEmailData struct {
	FromDate       time.Time `json:"from_date" db:"from_date"`
	ToDate         time.Time `json:"to_date" db:"to_date"`
	ServiceName    string    `json:"service_name" db:"service_name"`
	ShortLocation  string    `json:"short_location" db:"short_location"`
	CustomerEmail  string    `json:"customer_email" db:"customer_email"`
	EmailId        uuid.UUID `json:"email_id" db:"email_id"`
	MerchantName   string    `json:"merchant_name" db:"merchant_name"`
	CancelDeadline int       `json:"cancel_deadline" db:"cancel_deadline"`
}

func (s *service) GetBookingDataForEmail(ctx context.Context, bookingId int) (BookingEmailData, error) {
	query := `
	select b.from_date, b.to_date, bd.email_id, s.name as service_name, coalesce(u.email, c.email) as customer_email, m.name as merchant_name,
	coalesce(s.cancel_deadline, m.cancel_deadline) as cancel_deadline,
		l.address || ', ' || l.city || ', ' || l.postal_code || ', ' || l.country as short_location
	from "Booking" b
	join "Service" s on s.id = b.service_id
	join "BookingParticipant" bp on bp.booking_id = b.id
	join "Customer" c on c.id = bp.customer_id
	join "BookingDetails" bd on bd.booking_id = b.id
	left join "User" u on u.id = c.user_id
	join "Merchant" m on m.id = b.merchant_id
	join "Location" l on l.id = b.location_id
	where b.id = $1
	`

	var data BookingEmailData
	err := s.db.QueryRow(ctx, query, bookingId).Scan(&data.FromDate, &data.ToDate, &data.EmailId, &data.ServiceName,
		&data.CustomerEmail, &data.MerchantName, &data.CancelDeadline, &data.ShortLocation)
	if err != nil {
		return BookingEmailData{}, err
	}

	return data, nil
}

type PublicBookingInfo struct {
	FromDate       time.Time                `json:"from_date" db:"from_date"`
	ToDate         time.Time                `json:"to_date" db:"to_date"`
	ServiceName    string                   `json:"service_name" db:"service_name"`
	CancelDeadline int                      `json:"cancel_deadline" db:"cancel_deadline"`
	ShortLocation  string                   `json:"short_location" db:"short_location"`
	Price          currencyx.FormattedPrice `json:"price" db:"price"`
	PriceNote      *string                  `json:"price_note"`
	MerchantName   string                   `json:"merchant_name" db:"merchant_name"`
	IsCancelled    bool                     `json:"is_cancelled" db:"is_cancelled"`
}

func (s *service) GetPublicBookingInfo(ctx context.Context, bookingId int) (PublicBookingInfo, error) {
	query := `
	select b.from_date, b.to_date, bd.price_per_person as price, m.name as merchant_name, s.name as service_name, m.cancel_deadline, s.price_note,
		b.status = 'cancelled' as is_cancelled,
		l.address || ', ' || l.city || ' ' || l.postal_code || ', ' || l.country as short_location
	from "Booking" b
	join "BookingDetails" bd on bd.booking_id = b.id
	join "Service" s on s.id = b.service_id
	join "Merchant" m on m.id = b.merchant_id
	join "Location" l on l.id = b.location_id
	where b.id = $1
	`

	var data PublicBookingInfo
	err := s.db.QueryRow(ctx, query, bookingId).Scan(&data.FromDate, &data.ToDate, &data.Price, &data.MerchantName,
		&data.ServiceName, &data.CancelDeadline, &data.PriceNote, &data.IsCancelled, &data.ShortLocation)
	if err != nil {
		return PublicBookingInfo{}, err
	}

	return data, nil
}
