package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bojanz/currency"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/types"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
)

type Merchant struct {
	Id               uuid.UUID     `json:"ID"`
	Name             string        `json:"name"`
	UrlName          string        `json:"url_name"`
	ContactEmail     string        `json:"contact_email"`
	Introduction     string        `json:"introduction"`
	Announcement     string        `json:"announcement"`
	AboutUs          string        `json:"about_us"`
	ParkingInfo      string        `json:"parking_info"`
	PaymentInfo      string        `json:"payment_info"`
	Timezone         string        `json:"timezone"`
	CurrencyCode     string        `json:"currency_code"`
	SubscriptionTier types.SubTier `json:"subscription_tier"`
}

type Employee struct {
	Id          int                `json:"id"`
	UserId      *uuid.UUID         `json:"user_id"`
	MerchantId  uuid.UUID          `json:"merchant_id"`
	Role        types.EmployeeRole `json:"employee_role"`
	FirstName   *string            `json:"first_name"`
	LastName    *string            `json:"last_name"`
	Email       *string            `json:"email"`
	PhoneNumber *string            `json:"phone_number"`
	IsActive    bool               `json:"is_active"`
	InvitedOn   *time.Time         `json:"invited_on"`
	AcceptedOn  *time.Time         `json:"accpeted_on"`
}

type EmployeeLocation struct {
	EmployeeId int  `json:"employee_id"`
	LocationId int  `json:"location_id"`
	IsPrimary  bool `json:"is_primary"`
}

func (s *service) NewMerchant(ctx context.Context, userId uuid.UUID, merchant Merchant) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer tx.Rollback(ctx)

	newMerchantQuery := `
	insert into "Merchant" (ID, name, url_name, contact_email, introduction, announcement, about_us, parking_info,
		payment_info, timezone, currency_code, subscription_tier)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = tx.Exec(ctx, newMerchantQuery, merchant.Id, merchant.Name, merchant.UrlName, merchant.ContactEmail,
		merchant.Introduction, merchant.Announcement, merchant.AboutUs, merchant.ParkingInfo, merchant.PaymentInfo,
		merchant.Timezone, merchant.CurrencyCode, merchant.SubscriptionTier)
	if err != nil {
		return err
	}

	newPreferencesQuery := `
	insert into "Preferences" (merchant_id) values ($1)
	`

	_, err = tx.Exec(ctx, newPreferencesQuery, merchant.Id)
	if err != nil {
		return err
	}

	newEmployeeQuery := `
	insert into "Employee" (user_id, merchant_id, role, is_active)
	values ($1, $2, $3, $4)
	`

	_, err = tx.Exec(ctx, newEmployeeQuery, userId, merchant.Id, types.EmployeeRoleOwner, true)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *service) DeleteMerchant(ctx context.Context, employeeId int, merchantId uuid.UUID) error {
	query := `
	delete from "Merchant" m
	using "Employee" e
	where e.user_id = $1 and e.role = 'owner' and m.id = $2
	`

	_, err := s.db.Exec(ctx, query, employeeId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetMerchantIdByUrlName(ctx context.Context, UrlName string) (uuid.UUID, error) {
	query := `
	select id from "Merchant"
	where url_name = $1
	`

	var merchantId uuid.UUID
	err := s.db.QueryRow(ctx, query, UrlName).Scan(&merchantId)
	if err != nil {
		return uuid.Nil, err
	}

	return merchantId, nil
}

type MerchantInfo struct {
	Name         string `json:"merchant_name"`
	UrlName      string `json:"url_name"`
	ContactEmail string `json:"contact_email"`
	Introduction string `json:"introduction"`
	Announcement string `json:"announcement"`
	AboutUs      string `json:"about_us"`
	ParkingInfo  string `json:"parking_info"`
	PaymentInfo  string `json:"payment_info"`
	Timezone     string `json:"timezone"`

	LocationId        int            `json:"location_id"`
	Country           *string        `json:"country"`
	City              *string        `json:"city"`
	PostalCode        *string        `json:"postal_code"`
	Address           *string        `json:"address"`
	FormattedLocation string         `json:"formatted_location"`
	GeoPoint          types.GeoPoint `json:"geo_point"`

	Services []MerchantPageServicesGroupedByCategory `json:"services"`

	BusinessHours map[int][]TimeSlot `json:"business_hours"`
}

// TODO: this should be refactored ideally to one query
func (s *service) GetAllMerchantInfo(ctx context.Context, merchantId uuid.UUID) (MerchantInfo, error) {
	query := `
	select m.name, m.url_name, m.contact_email, m.introduction, m.announcement, m.about_us, m.parking_info, m.payment_info, m.timezone,
	l.id as location_id, l.country, l.city, l.postal_code, l.address, l.formatted_location, l.geo_point from "Merchant" m
	inner join "Location" l on m.id = l.merchant_id
	where m.id = $1
	`

	var mi MerchantInfo
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&mi.Name, &mi.UrlName, &mi.ContactEmail, &mi.Introduction, &mi.Announcement,
		&mi.AboutUs, &mi.ParkingInfo, &mi.PaymentInfo, &mi.Timezone, &mi.LocationId, &mi.Country, &mi.City, &mi.PostalCode, &mi.Address,
		&mi.FormattedLocation, &mi.GeoPoint)
	if err != nil {
		return MerchantInfo{}, err
	}

	mi.Services, err = s.GetServicesForMerchantPage(ctx, merchantId)
	if err != nil {
		return MerchantInfo{}, err
	}

	businnessHours, err := s.GetBusinessHours(ctx, merchantId)
	if err != nil {
		return MerchantInfo{}, fmt.Errorf("failed to get business hours for merchant: %v", err)
	}

	mi.BusinessHours = businnessHours

	return mi, nil
}

func (s *service) IsMerchantUrlUnique(ctx context.Context, merchantUrl string) error {
	query := `
	select 1 from "Merchant"
	where url_name = $1
	`

	var url string
	err := s.db.QueryRow(ctx, query, merchantUrl).Scan(&url)
	if !errors.Is(err, pgx.ErrNoRows) {
		if err != nil {
			return err
		}

		return fmt.Errorf("this merchant url is already used: %s", merchantUrl)
	}

	return nil
}

type PublicCustomer struct {
	Customer
	IsDummy         bool    `json:"is_dummy" db:"is_dummy"`
	IsBlacklisted   bool    `json:"is_blacklisted" db:"is_blacklisted"`
	BlacklistReason *string `json:"blacklist_reason" db:"blacklist_reason"`
	TimesBooked     int     `json:"times_booked" db:"times_booked"`
	TimesCancelled  int     `json:"times_cancelled" db:"times_cancelled"`
}

func (s *service) GetCustomersByMerchantId(ctx context.Context, merchantId uuid.UUID, isBlacklisted bool) ([]PublicCustomer, error) {
	query := `
	select c.id,
		   coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name,
		   coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number, c.birthday, c.note,
		   c.user_id is null as is_dummy, c.is_blacklisted, c.blacklist_reason,
		count(b.id) as times_booked, count(distinct bp.status = 'cancelled') as times_cancelled
	from "Customer" c
	left join "User" u on c.user_id = u.id
	left join "BookingParticipant" bp on c.id = coalesce(bp.transferred_to, bp.customer_id)
	left join "Booking" b on bp.booking_id = b.id and b.merchant_id = $1
	where c.merchant_id = $1 and c.is_blacklisted = $2
	group by c.id, u.first_name, u.last_name, u.email, u.phone_number
	`

	rows, _ := s.db.Query(ctx, query, merchantId, isBlacklisted)
	customers, err := pgx.CollectRows(rows, pgx.RowToStructByName[PublicCustomer])
	if err != nil {
		return []PublicCustomer{}, err
	}

	// if customers array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(customers) == 0 {
		customers = []PublicCustomer{}
	}

	return customers, nil
}

type TimeSlot struct {
	StartTime string `json:"start_time" db:"start_time"`
	EndTime   string `json:"end_time" db:"end_time"`
}

type MerchantSettingsInfo struct {
	Name             string             `json:"merchant_name" db:"merchant_name"`
	ContactEmail     string             `json:"contact_email" db:"contact_email"`
	Introduction     string             `json:"introduction" db:"introduction"`
	Announcement     string             `json:"announcement" db:"announcement"`
	AboutUs          string             `json:"about_us" db:"about_us"`
	ParkingInfo      string             `json:"parking_info" db:"parking_info"`
	PaymentInfo      string             `json:"payment_info" db:"payment_info"`
	CancelDeadline   int                `json:"cancel_deadline" db:"cancel_deadline"`
	BookingWindowMin int                `json:"booking_window_min" db:"booking_window_min"`
	BookingWindowMax int                `json:"booking_window_max" db:"booking_window_max"`
	BufferTime       int                `json:"buffer_time" db:"buffer_time"`
	Timezone         string             `json:"timezone" db:"timezone"`
	BusinessHours    map[int][]TimeSlot `json:"business_hours" db:"business_hours"`

	LocationId        int     `json:"location_id" db:"location_id"`
	Country           *string `json:"country" db:"country"`
	City              *string `json:"city" db:"city"`
	PostalCode        *string `json:"postal_code" db:"postal_code"`
	Address           *string `json:"address" db:"address"`
	FormattedLocation string  `json:"formatted_location" db:"formatted_location"`
}

func (s *service) GetMerchantSettingsInfo(ctx context.Context, merchantId uuid.UUID) (MerchantSettingsInfo, error) {

	var msi MerchantSettingsInfo

	merchantQuery := `
	select m.name, m.contact_email, m.introduction, m.announcement,
		   m.about_us, m.parking_info, m.payment_info, m.cancel_deadline, m.booking_window_min, m.booking_window_max, m.buffer_time, m.timezone,
	       l.id as location_id, l.country, l.city, l.postal_code, l.address, l.formatted_location
	from "Merchant" m inner join "Location" l on m.id = l.merchant_id
	where m.id = $1;`

	err := s.db.QueryRow(ctx, merchantQuery, merchantId).Scan(&msi.Name, &msi.ContactEmail, &msi.Introduction, &msi.Announcement,
		&msi.AboutUs, &msi.ParkingInfo, &msi.PaymentInfo, &msi.CancelDeadline, &msi.BookingWindowMin, &msi.BookingWindowMax, &msi.BufferTime, &msi.Timezone,
		&msi.LocationId, &msi.Country, &msi.City, &msi.PostalCode, &msi.Address, &msi.FormattedLocation)
	if err != nil {
		return MerchantSettingsInfo{}, err
	}

	businessHours, err := s.GetBusinessHours(ctx, merchantId)
	if err != nil {
		return MerchantSettingsInfo{}, fmt.Errorf("failed to get business hours for merchant: %v", err)
	}

	msi.BusinessHours = businessHours

	return msi, nil
}

type MerchantSettingFields struct {
	Introduction     string             `json:"introduction"`
	Announcement     string             `json:"announcement"`
	AboutUs          string             `json:"about_us"`
	ParkingInfo      string             `json:"parking_info"`
	PaymentInfo      string             `json:"payment_info"`
	CancelDeadline   int                `json:"cancel_deadline"`
	BookingWindowMin int                `json:"booking_window_min"`
	BookingWindowMax int                `json:"booking_window_max"`
	BufferTime       int                `json:"buffer_time"`
	BusinessHours    map[int][]TimeSlot `json:"business_hours"`
}

func (s *service) UpdateMerchantFieldsById(ctx context.Context, merchantId uuid.UUID, ms MerchantSettingFields) error {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer tx.Rollback(ctx)

	merchantQuery := `
	update "Merchant"
	set introduction = $2, announcement = $3, about_us = $4, payment_info = $5,
	parking_info = $6, cancel_deadline = $7, booking_window_min = $8, booking_window_max = $9, buffer_time = $10
	where id = $1;`

	_, err = tx.Exec(ctx, merchantQuery, merchantId, ms.Introduction, ms.Announcement, ms.AboutUs, ms.PaymentInfo, ms.ParkingInfo,
		ms.CancelDeadline, ms.BookingWindowMin, ms.BookingWindowMax, ms.BufferTime)
	if err != nil {
		return err
	}

	var days []int
	var starts, ends []string

	for day, timeRanges := range ms.BusinessHours {
		for _, ts := range timeRanges {
			if ts.StartTime != "" && ts.EndTime != "" {
				days = append(days, day)
				starts = append(starts, ts.StartTime)
				ends = append(ends, ts.EndTime)
			}
		}
	}

	deleteQuery := `
    delete from "BusinessHours"
    where merchant_id = $1
    and (day_of_week, start_time, end_time) not in (
        select unnest($2::int[]), unnest($3::time[]), unnest($4::time[])
    );`

	_, err = tx.Exec(ctx, deleteQuery, merchantId, utils.IntSliceToPgArray(days), utils.TimeStringToPgArray(starts), utils.TimeStringToPgArray(ends))
	if err != nil {
		return fmt.Errorf("failed to delete outdated business hours for merchant: %v", err)
	}

	insertQuery := `
    insert into "BusinessHours" (merchant_id, day_of_week, start_time, end_time)
    select $1, unnest($2::int[]), unnest($3::time[]), unnest($4::time[])
    on conflict (merchant_id, day_of_week, start_time, end_time) do nothing;`

	_, err = tx.Exec(ctx, insertQuery, merchantId, utils.IntSliceToPgArray(days), utils.TimeStringToPgArray(starts), utils.TimeStringToPgArray(ends))
	if err != nil {
		return fmt.Errorf("failed to insert business hours for merchant: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateBusinessHours(ctx context.Context, merchantId uuid.UUID, businessHours map[int][]TimeSlot) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer tx.Rollback(ctx)

	var days []int
	var starts, ends []string

	for day, timeRanges := range businessHours {
		for _, ts := range timeRanges {
			if ts.StartTime != "" && ts.EndTime != "" {
				days = append(days, day)
				starts = append(starts, ts.StartTime)
				ends = append(ends, ts.EndTime)
			}
		}
	}

	deleteQuery := `
    delete from "BusinessHours"
    where merchant_id = $1
    and (day_of_week, start_time, end_time) not in (
        select unnest($2::int[]), unnest($3::time[]), unnest($4::time[])
    );`

	_, err = tx.Exec(ctx, deleteQuery, merchantId, utils.IntSliceToPgArray(days), utils.TimeStringToPgArray(starts), utils.TimeStringToPgArray(ends))
	if err != nil {
		return fmt.Errorf("failed to delete outdated business hours for merchant: %v", err)
	}

	insertQuery := `
    insert into "BusinessHours" (merchant_id, day_of_week, start_time, end_time)
    select $1, unnest($2::int[]), unnest($3::time[]), unnest($4::time[])
    on conflict (merchant_id, day_of_week, start_time, end_time) do nothing;`

	_, err = tx.Exec(ctx, insertQuery, merchantId, utils.IntSliceToPgArray(days), utils.TimeStringToPgArray(starts), utils.TimeStringToPgArray(ends))
	if err != nil {
		return fmt.Errorf("failed to insert business hours for merchant: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *service) GetBusinessHours(ctx context.Context, merchantId uuid.UUID) (map[int][]TimeSlot, error) {
	query := `
	select day_of_week, start_time, end_time from "BusinessHours"
	where merchant_id = $1
	order by day_of_week, start_time;
	`

	rows, _ := s.db.Query(ctx, query, merchantId)

	businessHours := make(map[int][]TimeSlot)
	for day := 0; day <= 6; day++ {
		businessHours[day] = []TimeSlot{}
	}

	var dayOfWeek int
	var ts TimeSlot
	_, err := pgx.ForEachRow(rows, []any{&dayOfWeek, &ts.StartTime, &ts.EndTime}, func() error {
		ts.StartTime = strings.Split(ts.StartTime, ".")[0]
		ts.EndTime = strings.Split(ts.EndTime, ".")[0]

		businessHours[dayOfWeek] = append(businessHours[dayOfWeek], ts)

		return nil
	})
	if err != nil {
		return map[int][]TimeSlot{}, err
	}

	return businessHours, nil
}

func (s *service) GetBusinessHoursByDay(ctx context.Context, merchantId uuid.UUID, dayOfWeek int) ([]TimeSlot, error) {
	query := `
	select start_time, end_time from "BusinessHours"
	where merchant_id = $1 and day_of_week = $2
	order by start_time`

	rows, _ := s.db.Query(ctx, query, merchantId, dayOfWeek)
	bHours, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (TimeSlot, error) {
		var ts TimeSlot
		err := row.Scan(&ts.StartTime, &ts.EndTime)

		ts.StartTime = strings.Split(ts.StartTime, ".")[0]
		ts.EndTime = strings.Split(ts.EndTime, ".")[0]

		return ts, err
	})
	if err != nil {
		return []TimeSlot{}, err
	}

	return bHours, nil
}

func (s *service) GetNormalizedBusinessHours(ctx context.Context, merchantId uuid.UUID) (map[int]TimeSlot, error) {
	query := `
	select day_of_week, min(start_time) as start_time,
	max(end_time) as end_time from "BusinessHours"
	where merchant_id = $1
	group by day_of_week
	order by day_of_week;`

	rows, _ := s.db.Query(ctx, query, merchantId)

	var day int
	var startTime, endTime string

	result := make(map[int]TimeSlot)
	_, err := pgx.ForEachRow(rows, []any{&day, &startTime, &endTime}, func() error {
		startTime = strings.Split(startTime, ".")[0]
		endTime = strings.Split(endTime, ".")[0]

		result[day] = TimeSlot{
			StartTime: startTime,
			EndTime:   endTime,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *service) GetMerchantTimezoneById(ctx context.Context, merchantId uuid.UUID) (*time.Location, error) {
	query := `
	select timezone from "Merchant"
	where id = $1
	`

	var timzone string
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&timzone)
	if err != nil {
		return nil, err
	}

	tz, err := time.LoadLocation(timzone)
	if err != nil {
		return nil, fmt.Errorf("error while parsing merchant's timezone: %s", err.Error())
	}

	return tz, nil
}

type LowStockProduct struct {
	Id            int     `json:"id" db:"id"`
	Name          string  `json:"name" db:"name"`
	MaxAmount     int     `json:"max_amount" db:"max_amount"`
	CurrentAmount int     `json:"current_amount" db:"current_amount"`
	Unit          string  `json:"unit" db:"unit"`
	FillRatio     float64 `json:"fill_ratio" db:"fill_ratio"`
}

type DashboardData struct {
	PeriodStart      time.Time              `json:"period_start"`
	PeriodEnd        time.Time              `json:"period_end"`
	UpcomingBookings []PublicBookingDetails `json:"upcoming_bookings"`
	LatestBookings   []PublicBookingDetails `json:"latest_bookings"`
	LowStockProducts []LowStockProduct      `json:"low_stock_products"`
	Statistics       DashboardStatistics    `json:"statistics"`
}

func (s *service) GetDashboardData(ctx context.Context, merchantId uuid.UUID, date time.Time, period int) (DashboardData, error) {
	var dd DashboardData

	utcDate := date.UTC()

	// UpcomingBookings
	query := `
	select b.id, b.from_date, b.to_date, bp.customer_note, bd.merchant_note, bd.total_price as price, bd.total_cost as cost, s.name as service_name,
		s.color as service_color, s.total_duration as service_duration,
		coalesce(c.first_name, u.first_name) as first_name,
		coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	join "Service" s on b.service_id = s.id
	join "BookingDetails" bd on bd.booking_id = b.id
	join "BookingParticipant" bp on bp.booking_id = b.id
	left join "Customer" c on bp.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.status not in ('completed', 'cancelled') and bp.status not in ('completed', 'cancelled')
	order by b.from_date
	limit 5`

	var err error
	rows, _ := s.db.Query(ctx, query, merchantId, utcDate)
	dd.UpcomingBookings, err = pgx.CollectRows(rows, pgx.RowToStructByName[PublicBookingDetails])
	if err != nil {
		return DashboardData{}, err
	}

	// LatestBookings
	query2 := `
	select b.id, b.from_date, b.to_date, bp.customer_note, bd.merchant_note, bd.total_price as price, bd.total_cost as cost, s.name as service_name,
		s.color as service_color, s.total_duration as service_duration,
		coalesce(c.first_name, u.first_name) as first_name,
		coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.phone_number, u.phone_number) as phone_number
	from "Booking" b
	join "Service" s on b.service_id = s.id
	join "BookingDetails" bd on bd.booking_id = b.id
	join "BookingParticipant" bp on bp.booking_id = b.id
	left join "Customer" c on bp.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where b.merchant_id = $1 and b.from_date >= $2 AND b.status not in ('completed', 'cancelled') and bp.status not in ('completed', 'cancelled')
	order by b.id desc
	limit 5`

	rows, _ = s.db.Query(ctx, query2, merchantId, utcDate)
	dd.LatestBookings, err = pgx.CollectRows(rows, pgx.RowToStructByName[PublicBookingDetails])
	if err != nil {
		return DashboardData{}, err
	}

	// LowStockProducts
	query3 := `
	select p.id, p.name, p.max_amount, p.current_amount, p.unit, (p.current_amount::float / p.max_amount) as fill_ratio from "Product" p
	where  p.merchant_id = $1 and p.deleted_on is null and p.max_amount > 0 and (p.current_amount::float / p.max_amount) < 0.4
	order by fill_ratio asc`

	rows, _ = s.db.Query(ctx, query3, merchantId)
	dd.LowStockProducts, err = pgx.CollectRows(rows, pgx.RowToStructByName[LowStockProduct])
	if err != nil {
		return DashboardData{}, err
	}

	// -1 because the last is the current day
	currPeriodStart := utils.TruncateToDay(utcDate.AddDate(0, 0, -(period - 1)))
	prevPeriodStart := utils.TruncateToDay(currPeriodStart.AddDate(0, 0, -(period - 1)))

	dd.PeriodStart = currPeriodStart
	dd.PeriodEnd = utils.TruncateToDay(utcDate)

	dd.Statistics, err = s.getDashboardStatistics(ctx, merchantId, utcDate, currPeriodStart, prevPeriodStart)
	if err != nil {
		return DashboardData{}, err
	}

	return dd, nil
}

// TODO: value is of numeric type so float might not be the best
// type to return here
type RevenueStat struct {
	Value float64   `json:"value" db:"value"`
	Day   time.Time `json:"day" db:"day"`
}

type DashboardStatistics struct {
	Revenue               []RevenueStat `json:"revenue"`
	RevenueSum            string        `json:"revenue_sum"`
	RevenueChange         int           `json:"revenue_change"`
	Bookings              int           `json:"bookings"`
	BookingsChange        int           `json:"bookings_change"`
	Cancellations         int           `json:"cancellations"`
	CancellationsChange   int           `json:"cancellations_change"`
	AverageDuration       int           `json:"average_duration"`
	AverageDurationChange int           `json:"average_duration_change"`
}

// TODO: currently this assumes that all of the prices are/were in the same currency
func (s *service) getDashboardStatistics(ctx context.Context, merchantId uuid.UUID, date, currPeriodStart, prevPeriodStart time.Time) (DashboardStatistics, error) {
	query := `
	WITH base AS (
    SELECT
        to_date,
        (bd.total_price).number as price,
		(bd.total_price).currency as currency,
        EXTRACT(EPOCH FROM (to_date - from_date)) / 60 AS duration,
		(bp.status in ('cancelled')) as cancelled_by_user,
        (b.status in ('cancelled')) as cancelled
    FROM "Booking" b
	join "BookingDetails" bd on bd.booking_id = b.id
	join "BookingParticipant" bp on bp.booking_id = b.id
    WHERE merchant_id = $1
	order by b.id
	),
	current AS (
		SELECT
			SUM(price) FILTER (WHERE NOT cancelled) AS revenue,
			COUNT(*) FILTER (WHERE NOT cancelled) AS bookings,
			COUNT(*) FILTER (WHERE cancelled_by_user) AS cancellations,
			AVG(duration) FILTER (WHERE NOT cancelled) AS avg_duration
		FROM base
		WHERE to_date >= $2 AND to_date < $3
	),
	current_totals AS (
		SELECT
			COALESCE(SUM(revenue), 0) AS revenue_sum,
			COALESCE(SUM(bookings), 0) AS bookings,
			COALESCE(SUM(cancellations), 0) AS cancellations,
			COALESCE(CAST(AVG(avg_duration) AS INTEGER), 0) AS average_duration
		FROM current
	),
	previous AS (
		SELECT
			COALESCE(SUM(price) FILTER (WHERE NOT cancelled), 0) AS revenue_sum,
			COUNT(*) FILTER (WHERE NOT cancelled) AS bookings,
			COUNT(*) FILTER (WHERE cancelled_by_user) AS cancellations,
			CAST(AVG(duration) FILTER (WHERE NOT cancelled) AS INTEGER) AS average_duration
		FROM base
		WHERE to_date >= $4 AND to_date < $5
	),
	single_currency as (
		select currency
		from base
		group by currency
		order by count(*) desc, currency asc
		limit 1
	)
	SELECT
		ct.revenue_sum, p.revenue_sum,
		ct.bookings, p.bookings,
		ct.cancellations, p.cancellations,
		COALESCE(ct.average_duration, 0), COALESCE(p.average_duration, 0),
		sc.currency
	FROM current_totals ct
	cross join previous p
	cross join single_currency sc
	`

	var stats DashboardStatistics

	var (
		currRevenue, prevRevenue             int
		currBookings, prevBookings           int
		currCancellations, prevCancellations int
		currAvgDuration, prevAvgDuration     int
		curr                                 string
	)

	err := s.db.QueryRow(ctx, query, merchantId, currPeriodStart, date, prevPeriodStart, currPeriodStart.AddDate(0, 0, 1)).Scan(
		&currRevenue, &prevRevenue,
		&currBookings, &prevBookings,
		&currCancellations, &prevCancellations,
		&currAvgDuration, &prevAvgDuration,
		&curr,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return DashboardStatistics{}, err
		}
	}

	var formattedRevenue string

	// if no rows are returned
	if curr != "" {
		amount, err := currency.NewAmount(strconv.Itoa(currRevenue), curr)
		if err != nil {
			return DashboardStatistics{}, fmt.Errorf("new amount creation failed: %v", err)
		}
		formattedRevenue = currencyx.Format(amount)
	} else {
		formattedRevenue = "0"
	}

	stats.RevenueSum = formattedRevenue
	stats.Bookings = currBookings
	stats.Cancellations = currCancellations
	stats.AverageDuration = currAvgDuration

	stats.RevenueChange = utils.CalculatePercentChange(prevRevenue, currRevenue)
	stats.BookingsChange = utils.CalculatePercentChange(prevBookings, currBookings)
	stats.CancellationsChange = utils.CalculatePercentChange(prevCancellations, currCancellations) * -1
	stats.AverageDurationChange = utils.CalculatePercentChange(prevAvgDuration, currAvgDuration)

	// TODO: in the future only completed bookings should count towards revenue
	query2 := `
	SELECT
		DATE(bookings.from_date) AS day,
		COALESCE(SUM(bookings.price), 0) AS value
	FROM (
		select b.from_date, (bd.total_price).number as price
		from "Booking" b
		join "BookingDetails" bd on bd.booking_id = b.id
		where b.merchant_id = $1 AND b.from_date >= $2 AND b.from_date < $3 and b.status not in ('cancelled')
		order by b.id
	) as bookings
	GROUP BY day
	ORDER BY day
	`

	rows, _ := s.db.Query(ctx, query2, merchantId, currPeriodStart, date)
	stats.Revenue, err = pgx.CollectRows(rows, pgx.RowToStructByName[RevenueStat])
	if err != nil {
		return DashboardStatistics{}, err
	}

	return stats, nil
}

func (s *service) GetMerchantCurrency(ctx context.Context, merchantId uuid.UUID) (string, error) {
	query := `
	select currency_code from "Merchant" where id = $1
	`

	var curr string
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&curr)
	if err != nil {
		return "", err
	}

	return curr, nil
}

func (s *service) GetMerchantSubscriptionTier(ctx context.Context, merchantId uuid.UUID) (types.SubTier, error) {
	query := `
	select subscription_tier from "Merchant"
	where id = $1
	`

	var tier types.SubTier
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&tier)
	if err != nil {
		return types.SubTier{}, err
	}

	return tier, nil
}

type MerchantBookingSettings struct {
	BookingWindowMin int `json:"booking_window_min" db:"booking_window_min"`
	BookingWindowMax int `json:"booking_window_max" db:"booking_window_max"`
	BufferTime       int `json:"buffer_time" db:"buffer_time"`
}

func (s *service) GetBookingSettingsByMerchantAndService(ctx context.Context, merchantId uuid.UUID, serviceId int) (MerchantBookingSettings, error) {
	query := `
	select coalesce(s.buffer_time, m.buffer_time) as buffer_time,
	       coalesce(s.booking_window_max, m.booking_window_max) as booking_window_max,
		   coalesce(s.booking_window_min, m.booking_window_min) as booking_window_min
	from "Merchant" m
	join "Service" s on s.merchant_id = $1
	where m.id = $1 and s.id = $2`

	var mbs MerchantBookingSettings
	err := s.db.QueryRow(ctx, query, merchantId, serviceId).Scan(&mbs.BufferTime, &mbs.BookingWindowMax, &mbs.BookingWindowMin)
	if err != nil {
		return MerchantBookingSettings{}, err
	}

	return mbs, nil
}

func (s *service) ChangeMerchantNameAndURL(ctx context.Context, merchantId uuid.UUID, MerchantName, urlName string) error {
	query := `
	update "Merchant"
	set name = $2, url_name = $3
	where id = $1`

	_, err := s.db.Exec(ctx, query, merchantId, MerchantName, urlName)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) GetMerchantUrlNameById(ctx context.Context, merchantId uuid.UUID) (string, error) {
	query := `
	select url_name
	from "Merchant"
	where id = $1
	`

	var urlName string
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&urlName)
	if err != nil {
		return "", err
	}

	return urlName, nil
}

type BlockedTime struct {
	Id            int       `json:"id"`
	MerchantId    uuid.UUID `json:"merchant_id"`
	EmployeeId    int       `json:"employee_id"`
	BlockedTypeId *int      `json:"blocked_type_id"`
	Name          string    `json:"name"`
	FromDate      time.Time `json:"from_date"`
	ToDate        time.Time `json:"to_date"`
	AllDay        bool      `json:"all_day"`
}

func (s *service) NewBlockedTime(ctx context.Context, merchantId uuid.UUID, employeeIds []int, name string, fromDate, toDate time.Time, allDay bool, blockedTypeId *int) error {

	query := `
	insert into "BlockedTime" (merchant_id, employee_id, blocked_type_id, name, from_date, to_date, all_day) values ($1, $2, $3, $4, $5, $6, $7)`

	for _, empId := range employeeIds {
		_, err := s.db.Exec(ctx, query, merchantId, empId, blockedTypeId, name, fromDate, toDate, allDay)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *service) DeleteBlockedTime(ctx context.Context, blockedTimeId int, merchantId uuid.UUID, employeeId int) error {
	query := `
	delete from "BlockedTime"
	where merchant_id = $1 and employee_id = $2 and ID = $3`

	_, err := s.db.Exec(ctx, query, merchantId, employeeId, blockedTimeId)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateBlockedTime(ctx context.Context, bt BlockedTime) error {
	query := `
	update "BlockedTime"
	set blocked_type_id = $4, name = $5, from_date = $6, to_date = $7, all_day = $8 
	where merchant_id = $1 and employee_id = $2 and ID = $3`

	_, err := s.db.Exec(ctx, query, bt.MerchantId, bt.EmployeeId, bt.Id, bt.BlockedTypeId, bt.Name, bt.FromDate, bt.ToDate, bt.AllDay)
	if err != nil {
		return err
	}

	return nil
}

type BlockedTimes struct {
	FromDate time.Time `db:"from_date"`
	ToDate   time.Time `db:"to_date"`
	AllDay   bool      `db:"all_day"`
}

func (s *service) GetBlockedTimes(ctx context.Context, merchantId uuid.UUID, start, end time.Time) ([]BlockedTimes, error) {
	query := `
	select from_date, to_date, all_day from "BlockedTime"
	where merchant_id = $1 and to_date > $2 and from_date < $3
	order by from_date`

	rows, _ := s.db.Query(ctx, query, merchantId, start, end)
	blockedTimes, err := pgx.CollectRows(rows, pgx.RowToStructByName[BlockedTimes])
	if err != nil {
		return nil, err
	}

	return blockedTimes, nil

}

type EmployeeForCalendar struct {
	Id        int    `json:"id" db:"id"`
	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
}

func (s *service) GetEmployeesByMerchant(ctx context.Context, merchantId uuid.UUID) ([]EmployeeForCalendar, error) {
	query := `
	select e.id, coalesce(e.first_name, u.first_name) as first_name, coalesce(e.last_name, u.last_name) as last_name
	 from "Employee" e
	left join "User" u on u.id = e.user_id
	where merchant_id = $1`

	rows, _ := s.db.Query(ctx, query, merchantId)
	employees, err := pgx.CollectRows(rows, pgx.RowToStructByName[EmployeeForCalendar])
	if err != nil {
		return []EmployeeForCalendar{}, err
	}

	return employees, nil
}

type BlockedTimeType struct {
	Id       int    `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Duration int    `json:"duration" db:"duration"`
	Icon     string `json:"icon" db:"icon"`
}

func (s *service) GetAllBlockedTimeTypes(ctx context.Context, merchantId uuid.UUID) ([]BlockedTimeType, error) {
	query := `
	select ID, name, duration, icon from "BlockedTimeType"
	where merchant_id = $1
	order by id asc
	`

	rows, _ := s.db.Query(ctx, query, merchantId)
	types, err := pgx.CollectRows(rows, pgx.RowToStructByName[BlockedTimeType])
	if err != nil {
		return []BlockedTimeType{}, err
	}

	return types, nil
}

func (s *service) NewBlockedTimeType(ctx context.Context, merchantId uuid.UUID, btt BlockedTimeType) error {
	query := `
	insert into "BlockedTimeType" (merchant_id, name, duration, icon)
	values ($1, $2, $3, $4)
	`

	_, err := s.db.Exec(ctx, query, merchantId, btt.Name, btt.Duration, btt.Icon)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) UpdateBlockedTimeType(ctx context.Context, merchantId uuid.UUID, btt BlockedTimeType) error {
	query := `
	update "BlockedTimeType"
	set name = $3, duration = $4, icon = $5
	where merchant_id = $1 and id = $2
	`
	_, err := s.db.Exec(ctx, query, merchantId, btt.Id, btt.Name, btt.Duration, btt.Icon)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteBlockedTimeType(ctx context.Context, merchantId uuid.UUID, typeId int) error {
	query := `
	delete from "BlockedTimeType" where merchant_id = $1 and id = $2`

	_, err := s.db.Exec(ctx, query, merchantId, typeId)
	if err != nil {
		return err
	}

	return nil
}
