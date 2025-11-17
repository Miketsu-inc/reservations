package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/booking"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type Booking struct {
	Id                 int            `json:"ID" db:"id"`
	Status             booking.Status `json:"status" db:"status"`
	BookingType        booking.Type   `json:"booking_type" db:"booking_type"`
	IsRecurring        bool           `json:"is_recurring" db:"is_recurring"`
	MerchantId         uuid.UUID      `json:"merchant_id" db:"merchant_id"`
	EmployeeId         *int           `json:"employee_id" db:"employee_id"`
	ServiceId          int            `json:"service_id" db:"service_id"`
	LocationId         int            `json:"location_id" db:"location_id"`
	BookingSeriesId    int            `json:"booking_series_id" db:"booking_series_id"`
	SeriesOriginalDate time.Time      `json:"series_original_date" db:"series_original_date"`
	FromDate           time.Time      `json:"from_date" db:"from_date"`
	ToDate             time.Time      `json:"to_date" db:"to_date"`
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

type BookingSeries struct {
	Id          int          `json:"id" db:"id"`
	BookingType booking.Type `json:"booking_type" db:"booking_type"`
	MerchantId  uuid.UUID    `json:"merchant_id" db:"merchant_id"`
	EmployeeId  int          `json:"employee_id" db:"employee_id"`
	ServiceId   int          `json:"service_id" db:"service_id"`
	LocationId  int          `json:"location_id" db:"location_id"`
	Rrule       string       `json:"rrule" db:"rrule"`
	Dstart      time.Time    `json:"dstart" db:"dstart"`
	Timezone    string       `json:"timezone" db:"timezone"`
	IsActive    bool         `json:"is_active" db:"is_active"`
}

type BookingSeriesDetails struct {
	Id                  int             `json:"id" db:"id"`
	BookingSeriesId     int             `json:"booking_series_id" db:"booking_series_id"`
	PricePerPerson      currencyx.Price `json:"price_per_person" db:"price_per_person"`
	CostPerPerson       currencyx.Price `json:"cost_per_person" db:"cost_per_person"`
	TotalPrice          currencyx.Price `json:"total_price" db:"total_price"`
	TotalCost           currencyx.Price `json:"total_cost" db:"total_cost"`
	MinParticipants     int             `json:"min_participants" db:"min_participants"`
	MaxParticipants     int             `json:"max_participants" db:"max_participants"`
	CurrentParticipants int             `json:"current_participants" db:"current_participants"`
}

type BookingSeriesParticipant struct {
	Id              int        `json:"id" db:"id"`
	BookingSeriesId int        `json:"booking_series_id" db:"booking_series_id"`
	CustomerId      uuid.UUID  `json:"customer_id" db:"customer_id"`
	IsActive        bool       `json:"is_active" db:"is_active"`
	DroppedOutOn    *time.Time `json:"dropped_out_on" db:"dropped_out_on"`
}

type newBookingData struct {
	Status         booking.Status       `json:"status"`
	BookingType    booking.Type         `json:"booking_type"`
	MerchantId     uuid.UUID            `json:"merchant_id"`
	ServiceId      int                  `json:"service_id"`
	LocationId     int                  `json:"location_id"`
	FromDate       time.Time            `json:"from_date"`
	ToDate         time.Time            `json:"to_date"`
	CustomerNote   *string              `json:"customer_note"`
	MerchantNote   *string              `json:"merchant_note"`
	PricePerPerson currencyx.Price      `json:"price_per_person"`
	CostPerPerson  currencyx.Price      `json:"cost_per_person"`
	CustomerId     uuid.UUID            `json:"customer_id"`
	Phases         []PublicServicePhase `json:"phases"`
}

func newBooking(ctx context.Context, tx pgx.Tx, nb newBookingData) (int, error) {
	var bookingId int

	if nb.BookingType == booking.Appointment {
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

		_, err = tx.Exec(ctx, insertBookingDetailsQuery, bookingId, nb.PricePerPerson, nb.CostPerPerson, nb.PricePerPerson, nb.CostPerPerson, nb.MerchantNote, 1, 1, 1)
		if err != nil {
			return 0, err
		}

		insertBookingParticipantQuery := `
		insert into "BookingParticipant" (status, booking_id, customer_id, customer_note)
		values ($1, $2, $3, $4)
		`

		_, err = tx.Exec(ctx, insertBookingParticipantQuery, nb.Status, bookingId, nb.CustomerId, nb.CustomerNote)
		if err != nil {
			return 0, err
		}

	} else {
		assert.Never("TODO: Booking events or classes are not implemented yet!", nb)
	}

	return bookingId, nil
}

type NewCustomerBooking struct {
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

func (s *service) NewBookingByCustomer(ctx context.Context, nb NewCustomerBooking) (int, error) {
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

	bookingId, err := newBooking(ctx, tx, newBookingData{
		Status:         nb.Status,
		BookingType:    nb.BookingType,
		MerchantId:     nb.MerchantId,
		ServiceId:      nb.ServiceId,
		LocationId:     nb.LocationId,
		FromDate:       nb.FromDate,
		ToDate:         nb.ToDate,
		CustomerNote:   nb.CustomerNote,
		MerchantNote:   nil,
		PricePerPerson: nb.PricePerPerson,
		CostPerPerson:  nb.CostPerPerson,
		CustomerId:     customerId,
		Phases:         nb.Phases,
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

type NewMerchantBooking struct {
	Status         booking.Status       `json:"status"`
	BookingType    booking.Type         `json:"booking_type"`
	MerchantId     uuid.UUID            `json:"merchant_id"`
	ServiceId      int                  `json:"service_id"`
	LocationId     int                  `json:"location_id"`
	FromDate       time.Time            `json:"from_date"`
	ToDate         time.Time            `json:"to_date"`
	MerchantNote   *string              `json:"merchant_note"`
	PricePerPerson currencyx.Price      `json:"price_per_person"`
	CostPerPerson  currencyx.Price      `json:"cost_per_person"`
	CustomerId     uuid.UUID            `json:"customer_id"`
	Phases         []PublicServicePhase `json:"phases"`
}

func (s *service) NewBookingByMerchant(ctx context.Context, nb NewMerchantBooking) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	bookingId, err := newBooking(ctx, tx, newBookingData{
		Status:         nb.Status,
		BookingType:    nb.BookingType,
		MerchantId:     nb.MerchantId,
		ServiceId:      nb.ServiceId,
		LocationId:     nb.LocationId,
		FromDate:       nb.FromDate,
		ToDate:         nb.ToDate,
		CustomerNote:   nil,
		MerchantNote:   nb.MerchantNote,
		PricePerPerson: nb.PricePerPerson,
		CostPerPerson:  nb.CostPerPerson,
		CustomerId:     nb.CustomerId,
		Phases:         nb.Phases,
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
	CustomerNote    *string                  `json:"customer_note" db:"customer_note"`
	MerchantNote    *string                  `json:"merchant_note" db:"merchant_note"`
	ServiceName     string                   `json:"service_name" db:"service_name"`
	ServiceColor    string                   `json:"service_color" db:"service_color"`
	ServiceDuration int                      `json:"service_duration" db:"service_duration"`
	Price           currencyx.FormattedPrice `json:"price" db:"price"`
	Cost            currencyx.FormattedPrice `json:"cost" db:"cost"`
	FirstName       string                   `json:"first_name" db:"first_name"`
	LastName        string                   `json:"last_name" db:"last_name"`
	PhoneNumber     string                   `json:"phone_number" db:"phone_number"`
}

type BlockedTimeEvent struct {
	ID         int       `json:"id" db:"id"`
	EmployeeId int       `json:"employee_id" db:"employee_id"`
	Name       string    `json:"name" db:"name"`
	FromDate   time.Time `json:"from_date" db:"from_date"`
	ToDate     time.Time `json:"to_date" db:"to_date"`
	AllDay     bool      `json:"all_day" db:"all_day"`
}

type CalcendarEvents struct {
	Bookings     []PublicBookingDetails `json:"bookings"`
	BlockedTimes []BlockedTimeEvent     `json:"blocked_times"`
}

func (s *service) GetCalendarEventsByMerchant(ctx context.Context, merchantId uuid.UUID, start string, end string) (CalcendarEvents, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return CalcendarEvents{}, err
	}

	// nolint:errcheck
	defer tx.Rollback(ctx)

	var events CalcendarEvents

	bookingQuery := `
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
	rows, _ := tx.Query(ctx, bookingQuery, merchantId, start, end)
	events.Bookings, err = pgx.CollectRows(rows, pgx.RowToStructByName[PublicBookingDetails])
	if err != nil {
		return CalcendarEvents{}, err
	}

	blockedTimeQuery := `
	select id, employee_id, name, from_date, to_date, all_day from "BlockedTime"
	where merchant_id = $1 and ((from_date >= $2 and from_date <= $3) or (to_date >= $2 and to_date <= $3))
	order by id
	`

	rows, _ = tx.Query(ctx, blockedTimeQuery, merchantId, start, end)
	events.BlockedTimes, err = pgx.CollectRows(rows, pgx.RowToStructByName[BlockedTimeEvent])
	if err != nil {
		return CalcendarEvents{}, err
	}

	if err = tx.Commit(ctx); err != nil {
		return CalcendarEvents{}, err
	}

	return events, nil
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

type NewRecurringBookings struct {
	BookingSeriesId int
	BookingStatus   booking.Status
	BookingType     booking.Type
	MerchantId      uuid.UUID
	EmployeeId      int
	ServiceId       int
	LocationId      int
	FromDates       []time.Time
	ToDates         []time.Time
	Phases          []PublicServicePhase `json:"phases"`
	Details         BookingSeriesDetails
	Participants    []BookingSeriesParticipant
}

func (s *service) BatchCreateRecurringBookings(ctx context.Context, nrb NewRecurringBookings) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	assert.True(len(nrb.FromDates) == len(nrb.ToDates), "Length of fromDates and toDates slices should be equal!", len(nrb.FromDates), len(nrb.ToDates), nrb)
	var bookingIds []int

	if nrb.BookingType == booking.Appointment {
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

		_, err = tx.Exec(ctx, insertBookingPhasesQuery, bookingIds, phaseIds, phaseFromDates, phaseToDates)
		if err != nil {
			return 0, err
		}

		insertBookingDetailsQuery := `
		insert into "BookingDetails" (booking_id, price_per_person, cost_per_person, total_price, total_cost, merchant_note, min_participants, max_participants, current_participants)
		select unnest($1::int[]), $2, $3, $4, $5, $6, $7, $8, $9
		`

		_, err = tx.Exec(ctx, insertBookingDetailsQuery, bookingIds, nrb.Details.PricePerPerson, nrb.Details.CostPerPerson, nrb.Details.PricePerPerson, nrb.Details.CostPerPerson,
			"", nrb.Details.MinParticipants, nrb.Details.MaxParticipants, nrb.Details.CurrentParticipants)
		if err != nil {
			return 0, err
		}

		insertBookingParticipantsQuery := `
		insert into "BookingParticipant" (status, booking_id, customer_id, customer_note)
		select $1, unnest($2::int[]), $3, $4
		`

		_, err = tx.Exec(ctx, insertBookingParticipantsQuery, booking.Booked, bookingIds, nrb.Participants[0].CustomerId, "")
		if err != nil {
			return 0, err
		}

	} else {
		assert.Never("TODO: Booking events or classes are not implemented yet!", nrb)
	}

	return bookingIds[0], tx.Commit(ctx)
}

func (s *service) GetExistingOccurrenceDates(ctx context.Context, seriesId int, fromDate, toDate time.Time) ([]time.Time, error) {
	query := `
	select series_original_date
	from "Booking"
	where booking_series_id = $1 and series_original_date >= $2 and series_original_date <= $3
	`

	rows, _ := s.db.Query(ctx, query, seriesId, fromDate, toDate)
	dates, err := pgx.CollectRows(rows, pgx.RowTo[time.Time])
	if err != nil {
		return []time.Time{}, nil
	}

	return dates, nil
}

type CompleteBookingSeries struct {
	BookingSeries
	Details      BookingSeriesDetails
	Participants []BookingSeriesParticipant
}

type NewBookingSeries struct {
	BookingType    booking.Type
	MerchantId     uuid.UUID
	EmployeeId     int
	ServiceId      int
	LocationId     int
	Rrule          string
	Dstart         time.Time
	Timezone       *time.Location
	PricePerPerson currencyx.Price
	CostPerPerson  currencyx.Price
	Participants   []uuid.UUID
}

func (s *service) NewBookingSeries(ctx context.Context, nbs NewBookingSeries) (CompleteBookingSeries, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return CompleteBookingSeries{}, err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	var booking CompleteBookingSeries

	insertBookingSeriesQuery := `
	insert into "BookingSeries" (booking_type, merchant_id, employee_id, service_id, location_id, rrule, dstart, timezone, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	returning *
	`

	rows, _ := tx.Query(ctx, insertBookingSeriesQuery, nbs.BookingType, nbs.MerchantId, nbs.EmployeeId, nbs.ServiceId, nbs.LocationId, nbs.Rrule, nbs.Dstart,
		nbs.Timezone.String(), true)
	bookingSeries, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[BookingSeries])
	if err != nil {
		return CompleteBookingSeries{}, err
	}

	insertBookingSeriesDetailsQuery := `
	insert into "BookingSeriesDetails" (booking_series_id, price_per_person, cost_per_person, total_price, total_cost, min_participants, max_participants, current_participants)
	values ($1, $2, $3, $4, $5, $6, $7, $8)
	returning *
	`

	rows, _ = tx.Query(ctx, insertBookingSeriesDetailsQuery, bookingSeries.Id, nbs.PricePerPerson, nbs.CostPerPerson, nbs.PricePerPerson, nbs.CostPerPerson, 1, 1, 1)
	bookingSeriesDetails, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[BookingSeriesDetails])
	if err != nil {
		return CompleteBookingSeries{}, err
	}

	insertBookingSeriesParticipantsQuery := `
	insert into "BookingSeriesParticipant" (booking_series_id, customer_id, is_active)
	select $1, unnest($2::uuid[]), $3
	returning *
	`

	rows, _ = tx.Query(ctx, insertBookingSeriesParticipantsQuery, bookingSeries.Id, nbs.Participants, true)
	bookingSeriesParticipants, err := pgx.CollectRows(rows, pgx.RowToStructByName[BookingSeriesParticipant])
	if err != nil {
		return CompleteBookingSeries{}, err
	}

	booking.BookingSeries = bookingSeries
	booking.Details = bookingSeriesDetails
	booking.Participants = bookingSeriesParticipants

	err = tx.Commit(ctx)
	if err != nil {
		return CompleteBookingSeries{}, err
	}

	return booking, nil
}
