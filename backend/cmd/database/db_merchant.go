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
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/subscription"
)

type Merchant struct {
	Id               uuid.UUID         `json:"ID"`
	Name             string            `json:"name"`
	UrlName          string            `json:"url_name"`
	OwnerId          uuid.UUID         `json:"owner_id"`
	ContactEmail     string            `json:"contact_email"`
	Introduction     string            `json:"introduction"`
	Announcement     string            `json:"announcement"`
	AboutUs          string            `json:"about_us"`
	ParkingInfo      string            `json:"parking_info"`
	PaymentInfo      string            `json:"payment_info"`
	Timezone         string            `json:"timezone"`
	CurrencyCode     string            `json:"currency_code"`
	SubscriptionTier subscription.Tier `json:"subscription_tier"`
}

func (s *service) NewMerchant(ctx context.Context, merchant Merchant) error {
	query := `
	insert into "Merchant" (ID, name, url_name, owner_id, contact_email, introduction, announcement, about_us, parking_info,
		payment_info, timezone, currency_code, subscription_tier)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := s.db.Exec(ctx, query, merchant.Id, merchant.Name, merchant.UrlName, merchant.OwnerId, merchant.ContactEmail,
		merchant.Introduction, merchant.Announcement, merchant.AboutUs, merchant.ParkingInfo, merchant.PaymentInfo,
		merchant.Timezone, merchant.CurrencyCode, merchant.SubscriptionTier)
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

func (s *service) GetMerchantIdByOwnerId(ctx context.Context, ownerId uuid.UUID) (uuid.UUID, error) {
	query := `
	select id from "Merchant"
	where owner_id = $1
	`

	var merchantId uuid.UUID
	err := s.db.QueryRow(ctx, query, ownerId).Scan(&merchantId)
	if err != nil {
		return uuid.UUID{}, err
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

	LocationId int    `json:"location_id"`
	Country    string `json:"country"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Address    string `json:"address"`

	Services []MerchantPageServicesGroupedByCategory `json:"services"`

	BusinessHours map[int][]TimeSlot `json:"business_hours"`
}

// TODO: this should be refactored ideally to one query
func (s *service) GetAllMerchantInfo(ctx context.Context, merchantId uuid.UUID) (MerchantInfo, error) {
	query := `
	select m.name, m.url_name, m.contact_email, m.introduction, m.announcement, m.about_us, m.parking_info, m.payment_info, m.timezone,
	l.id as location_id, l.country, l.city, l.postal_code, l.address from "Merchant" m
	inner join "Location" l on m.id = l.merchant_id
	where m.id = $1
	`

	var mi MerchantInfo
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&mi.Name, &mi.UrlName, &mi.ContactEmail, &mi.Introduction, &mi.Announcement,
		&mi.AboutUs, &mi.ParkingInfo, &mi.PaymentInfo, &mi.Timezone, &mi.LocationId, &mi.Country, &mi.City, &mi.PostalCode, &mi.Address)
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

func (s *service) GetCustomersByMerchantId(ctx context.Context, merchantId uuid.UUID) ([]PublicCustomer, error) {
	query := `
	select c.id,
		   coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name,
		   coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number, c.birthday, c.note,
		   c.user_id is null as is_dummy, c.is_blacklisted, c.blacklist_reason,
		count(distinct a.group_id) as times_booked, count(distinct case when a.cancelled_by_user_on is not null then a.group_id end) as times_cancelled
	from "Customer" c
	left join "User" u on c.user_id = u.id
	left join "Appointment" a on (c.id = a.customer_id or a.transferred_to = c.id) and a.merchant_id = $1
	where c.merchant_id = $1 and c.is_blacklisted is false
	group by c.id, u.first_name, u.last_name, u.email, u.phone_number
	`

	rows, _ := s.db.Query(ctx, query, merchantId)
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

func (s *service) GetBlacklistedCustomersByMerchantId(ctx context.Context, merchantId uuid.UUID) ([]PublicCustomer, error) {
	query := `
	select c.id,
		   coalesce(c.first_name, u.first_name) as first_name, coalesce(c.last_name, u.last_name) as last_name,
		   coalesce(c.email, u.email) as email, coalesce(c.phone_number, u.phone_number) as phone_number,  c.birthday, c.note,
		   c.user_id is null as is_dummy, c.is_blacklisted, c.blacklist_reason,
		count(distinct a.group_id) as times_booked, count(distinct case when a.cancelled_by_user_on is not null then a.group_id end) as times_cancelled
	from "Customer" c
	left join "User" u on c.user_id = u.id
	left join "Appointment" a on (c.id = a.customer_id or a.transferred_to = c.id) and a.merchant_id = $1
	where c.merchant_id = $1 and c.is_blacklisted is true
	group by c.id, u.first_name, u.last_name, u.email, u.phone_number
	`

	rows, _ := s.db.Query(ctx, query, merchantId)
	customers, err := pgx.CollectRows(rows, pgx.RowToStructByName[PublicCustomer])
	if err != nil {
		return customers, nil
	}

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
	Name          string             `json:"merchant_name" db:"merchant_name"`
	ContactEmail  string             `json:"contact_email" db:"contact_email"`
	Introduction  string             `json:"introduction" db:"introduction"`
	Announcement  string             `json:"announcement" db:"announcement"`
	AboutUs       string             `json:"about_us" db:"about_us"`
	ParkingInfo   string             `json:"parking_info" db:"parking_info"`
	PaymentInfo   string             `json:"payment_info" db:"payment_info"`
	Timezone      string             `json:"timezone" db:"timezone"`
	BusinessHours map[int][]TimeSlot `json:"business_hours" db:"business_hours"`

	LocationId int    `json:"location_id" db:"location_id"`
	Country    string `json:"country" db:"country"`
	City       string `json:"city" db:"city"`
	PostalCode string `json:"postal_code" db:"postal_code"`
	Address    string `json:"address" db:"address"`
}

func (s *service) GetMerchantSettingsInfo(ctx context.Context, merchantId uuid.UUID) (MerchantSettingsInfo, error) {

	var msi MerchantSettingsInfo

	merchantQuery := `
	select m.name, m.contact_email, m.introduction, m.announcement,
		   m.about_us, m.parking_info, m.payment_info, m.timezone,
	       l.id as location_id, l.country, l.city, l.postal_code, l.address
	from "Merchant" m inner join "Location" l on m.id = l.merchant_id
	where m.id = $1;`

	err := s.db.QueryRow(ctx, merchantQuery, merchantId).Scan(&msi.Name, &msi.ContactEmail, &msi.Introduction, &msi.Announcement,
		&msi.AboutUs, &msi.ParkingInfo, &msi.PaymentInfo, &msi.Timezone, &msi.LocationId, &msi.Country, &msi.City, &msi.PostalCode, &msi.Address)
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

func (s *service) UpdateMerchantFieldsById(ctx context.Context, merchantId uuid.UUID, introduction, announcement, aboutUs, paymentInfo, parkingInfo string,
	businessHours map[int][]TimeSlot) error {

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer tx.Rollback(ctx)

	merchantQuery := `
	update "Merchant"
	set introduction = $2, announcement = $3, about_us = $4, payment_info = $5, parking_info = $6
	where id = $1;`

	_, err = tx.Exec(ctx, merchantQuery, merchantId, introduction, announcement, aboutUs, paymentInfo, parkingInfo)
	if err != nil {
		return err
	}

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

func (s *service) GetMerchantTimezoneById(ctx context.Context, merchantId uuid.UUID) (string, error) {
	query := `
	select timezone from "Merchant"
	where id = $1
	`

	var timzone string
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&timzone)
	if err != nil {
		return "", err
	}

	return timzone, nil
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
	PeriodStart          time.Time            `json:"period_start"`
	PeriodEnd            time.Time            `json:"period_end"`
	UpcomingAppointments []AppointmentDetails `json:"upcoming_appointments"`
	LatestBookings       []AppointmentDetails `json:"latest_bookings"`
	LowStockProducts     []LowStockProduct    `json:"low_stock_products"`
	Statistics           DashboardStatistics  `json:"statistics"`
}

func (s *service) GetDashboardData(ctx context.Context, merchantId uuid.UUID, date time.Time, period int) (DashboardData, error) {
	var dd DashboardData

	utcDate := date.UTC()

	// UpcomingAppointments
	query := `
	select distinct on (a.group_id) a.id, a.group_id,
		min(a.from_date) over (partition by a.group_id) as from_date,
		max(a.to_date) over (partition by a.group_id) as to_date,
		a.customer_note, a.merchant_note, a.price_then as price, a.cost_then as cost, s.name as service_name,
		s.color as service_color, s.total_duration as service_duration,
		coalesce(c.first_name, u.first_name) as first_name,
		coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.phone_number, u.phone_number) as phone_number
	from "Appointment" a
	join "Service" s on a.service_id = s.id
	join "Customer" c on a.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where a.merchant_id = $1 and a.from_date >= $2 AND a.cancelled_by_user_on is null and a.cancelled_by_merchant_on is null
	order by a.group_id, a.from_date
	limit 5`

	var err error
	rows, _ := s.db.Query(ctx, query, merchantId, utcDate)
	dd.UpcomingAppointments, err = pgx.CollectRows(rows, pgx.RowToStructByName[AppointmentDetails])
	if err != nil {
		return DashboardData{}, err
	}

	// LatestBookings
	query2 := `
	select distinct on (a.group_id) a.id, a.group_id,
		min(a.from_date) over (partition by a.group_id) as from_date,
		max(a.to_date) over (partition by a.group_id) as to_date,
		a.customer_note, a.merchant_note, a.price_then as price, a.cost_then as cost, s.name as service_name,
		s.color as service_color, s.total_duration as service_duration,
		coalesce(c.first_name, u.first_name) as first_name,
		coalesce(c.last_name, u.last_name) as last_name,
		coalesce(c.phone_number, u.phone_number) as phone_number
	from "Appointment" a
	join "Service" s on a.service_id = s.id
	join "Customer" c on a.customer_id = c.id
	left join "User" u on c.user_id = u.id
	where a.merchant_id = $1 and a.from_date >= $2 AND a.cancelled_by_user_on is null and a.cancelled_by_merchant_on is null
	order by a.group_id, a.id desc
	limit 5`

	rows, _ = s.db.Query(ctx, query2, merchantId, utcDate)
	dd.LatestBookings, err = pgx.CollectRows(rows, pgx.RowToStructByName[AppointmentDetails])
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
	Appointments          int           `json:"appointments"`
	AppointmentsChange    int           `json:"appointments_change"`
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
		distinct on (group_id)
        to_date,
        (price_then).number as price,
		(price_then).currency as currency,
        EXTRACT(EPOCH FROM (to_date - from_date)) / 60 AS duration,
        cancelled_by_user_on is not null as cancelled_by_user,
        (cancelled_by_user_on is not null or cancelled_by_merchant_on is not null) as cancelled
    FROM "Appointment" a
    WHERE merchant_id = $1
	order by group_id, id
	),
	current AS (
		SELECT
			SUM(price) FILTER (WHERE NOT cancelled) AS revenue,
			COUNT(*) FILTER (WHERE NOT cancelled) AS appointments,
			COUNT(*) FILTER (WHERE cancelled_by_user) AS cancellations,
			AVG(duration) FILTER (WHERE NOT cancelled) AS avg_duration
		FROM base
		WHERE to_date >= $2 AND to_date < $3
	),
	current_totals AS (
		SELECT
			COALESCE(SUM(revenue), 0) AS revenue_sum,
			COALESCE(SUM(appointments), 0) AS appointments,
			COALESCE(SUM(cancellations), 0) AS cancellations,
			COALESCE(CAST(AVG(avg_duration) AS INTEGER), 0) AS average_duration
		FROM current
	),
	previous AS (
		SELECT
			COALESCE(SUM(price) FILTER (WHERE NOT cancelled), 0) AS revenue_sum,
			COUNT(*) FILTER (WHERE NOT cancelled) AS appointments,
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
		ct.appointments, p.appointments,
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
		currAppointments, prevAppointments   int
		currCancellations, prevCancellations int
		currAvgDuration, prevAvgDuration     int
		curr                                 string
	)

	err := s.db.QueryRow(ctx, query, merchantId, currPeriodStart, date, prevPeriodStart, currPeriodStart.AddDate(0, 0, 1)).Scan(
		&currRevenue, &prevRevenue,
		&currAppointments, &prevAppointments,
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
	stats.Appointments = currAppointments
	stats.Cancellations = currCancellations
	stats.AverageDuration = currAvgDuration

	stats.RevenueChange = utils.CalculatePercentChange(prevRevenue, currRevenue)
	stats.AppointmentsChange = utils.CalculatePercentChange(prevAppointments, currAppointments)
	stats.CancellationsChange = utils.CalculatePercentChange(prevCancellations, currCancellations) * -1
	stats.AverageDurationChange = utils.CalculatePercentChange(prevAvgDuration, currAvgDuration)

	query2 := `
	SELECT
		DATE(from_date) AS day,
		COALESCE(SUM(price), 0) AS value
	FROM (
		select distinct on (group_id) from_date, (price_then).number as price
		from "Appointment"
		WHERE merchant_id = $1 AND from_date >= $2 AND from_date < $3
		AND cancelled_by_user_on IS NULL AND cancelled_by_merchant_on IS NULL
		order by group_id
	)
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

func (s *service) GetMerchantSubscriptionTier(ctx context.Context, merchantId uuid.UUID) (subscription.Tier, error) {
	query := `
	select subscription_tier from "Merchant"
	where id = $1
	`

	var tier subscription.Tier
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&tier)
	if err != nil {
		return subscription.Tier{}, err
	}

	return tier, nil
}
