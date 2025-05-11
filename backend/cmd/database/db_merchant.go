package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/cmd/utils"
)

type Merchant struct {
	Id           uuid.UUID `json:"ID"`
	Name         string    `json:"name"`
	UrlName      string    `json:"url_name"`
	OwnerId      uuid.UUID `json:"owner_id"`
	ContactEmail string    `json:"contact_email"`
	Introduction string    `json:"introduction"`
	Announcement string    `json:"announcement"`
	AboutUs      string    `json:"about_us"`
	ParkingInfo  string    `json:"parking_info"`
	PaymentInfo  string    `json:"payment_info"`
	Timezone     string    `json:"timezone"`
}

func (s *service) NewMerchant(ctx context.Context, merchant Merchant) error {
	query := `
	insert into "Merchant" (ID, name, url_name, owner_id, contact_email, introduction, announcement, about_us, parking_info, payment_info, timezone)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(ctx, query, merchant.Id, merchant.Name, merchant.UrlName, merchant.OwnerId, merchant.ContactEmail,
		merchant.Introduction, merchant.Announcement, merchant.AboutUs, merchant.ParkingInfo, merchant.PaymentInfo, merchant.Timezone)
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

func (s *service) GetMerchantById(ctx context.Context, merchantId uuid.UUID) (Merchant, error) {
	query := `
	select * from "Merchant"
	where id = $1
	`

	var merchant Merchant
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&merchant.Id, &merchant.Name, &merchant.UrlName, &merchant.OwnerId, &merchant.ContactEmail,
		&merchant.Introduction, &merchant.Announcement, &merchant.AboutUs, &merchant.ParkingInfo, &merchant.PaymentInfo)
	if err != nil {
		return Merchant{}, err
	}

	return merchant, nil
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

	Services []PublicService `json:"services"`

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

	mi.Services, err = s.GetServicesByMerchantId(ctx, merchantId)
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
	IsBlacklisted  bool `json:"is_blacklisted" db:"is_blacklisted"`
	TimesBooked    int  `json:"times_booked" db:"times_booked"`
	TimesCancelled int  `json:"times_cancelled" db:"times_cancelled"`
}

func (s *service) GetCustomersByMerchantId(ctx context.Context, merchantId uuid.UUID) ([]PublicCustomer, error) {
	query := `
	select u.id, u.first_name, u.last_name, u.email, u.phone_number, u.is_dummy, b.user_id is not null as is_blacklisted,
		count(a.id) as times_booked, count(case when a.cancelled_by_user_on is not null then 1 end) as times_cancelled
	from "User" u
	left join "Appointment" a on u.id = a.user_id and a.merchant_id = $1
	left join "Blacklist" b on u.id = b.user_id and b.merchant_id = $2
	where u.is_dummy = true or a.id is not null
	group by u.id, b.user_id;
	`

	rows, _ := s.db.Query(ctx, query, merchantId, merchantId)
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

func (s *service) UpdateMerchantFieldsById(ctx context.Context, merchantId uuid.UUID, introduction, announcement, aboutUs, paymentInfo, parkingInfo string, businessHours map[int][]TimeSlot) error {

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

	// 2. Insert new rows (avoiding duplicates)
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
	Id            int    `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	StockQuantity int    `json:"stock_quantity" db:"stock_quantity"`
	UsagePerUnit  int    `json:"usage_per_unit" db:"usage_per_unit"`
}

type DashboardData struct {
	UpcomingAppointments []AppointmentDetails `json:"upcoming_appointments"`
	LatestBookings       []AppointmentDetails `json:"latest_bookings"`
	LowStockProducts     []LowStockProduct    `json:"low_stock_products"`
	Statistics           DashboardStatistics  `json:"statistics"`
}

func (s *service) GetDashboardData(ctx context.Context, merchantId uuid.UUID, date time.Time, period int) (DashboardData, error) {
	var dd DashboardData

	utcDate := date.UTC()
	currPeriodStart := utcDate.AddDate(0, 0, -period)
	prevPeriodStart := currPeriodStart.AddDate(0, 0, -period)

	// UpcomingAppointments
	query := `
	select a.id, a.from_date, a.to_date, a.user_note, a.merchant_note, a.price_then as price, a.cost_then as cost,
	s.name as service_name, s.color as service_color, s.duration as service_duration, u.first_name, u.last_name, u.phone_number
	from "Appointment" a
	join "Service" s on a.service_id = s.id
	join "User" u on a.user_id = u.id
	where a.merchant_id = $1 and a.from_date >= $2 AND a.cancelled_by_user_on is null and a.cancelled_by_merchant_on is null
	order by a.from_date
	limit 5`

	var err error
	rows, _ := s.db.Query(ctx, query, merchantId, utcDate)
	dd.UpcomingAppointments, err = pgx.CollectRows(rows, pgx.RowToStructByName[AppointmentDetails])
	if err != nil {
		return DashboardData{}, err
	}

	// LatestBookings
	query2 := `
	select a.id, a.from_date, a.to_date, a.user_note, a.merchant_note, a.price_then as price, a.cost_then as cost,
	s.name as service_name, s.color as service_color, s.duration as service_duration, u.first_name, u.last_name, u.phone_number
	from "Appointment" a
	join "Service" s on a.service_id = s.id
	join "User" u on a.user_id = u.id
	where a.merchant_id = $1 and a.from_date >= $2 AND a.cancelled_by_user_on is null and a.cancelled_by_merchant_on is null
	order by a.id desc
	limit 5`

	rows, _ = s.db.Query(ctx, query2, merchantId, utcDate)
	dd.LatestBookings, err = pgx.CollectRows(rows, pgx.RowToStructByName[AppointmentDetails])
	if err != nil {
		return DashboardData{}, err
	}

	// LowStockProducts
	query3 := `
	select p.id, p.name, p.stock_quantity, p.usage_per_unit
	from "Product" p
	where p.merchant_id = $1 and p.stock_quantity < 5 and deleted_on is null
	order by p.stock_quantity
	`

	rows, _ = s.db.Query(ctx, query3, merchantId)
	dd.LowStockProducts, err = pgx.CollectRows(rows, pgx.RowToStructByName[LowStockProduct])
	if err != nil {
		return DashboardData{}, err
	}

	dd.Statistics, err = s.getDashboardStatistics(ctx, merchantId, utcDate, currPeriodStart, prevPeriodStart)
	if err != nil {
		return DashboardData{}, err
	}

	return dd, nil
}

type RevenueStat struct {
	Revenue int `json:"revenue"`
	Day     int `json:"day"`
}

type DashboardStatistics struct {
	Revenue               []RevenueStat `json:"revenue"`
	RevenueSum            int           `json:"revenue_sum"`
	RevenueChange         int           `json:"revenue_change"`
	Appointments          int           `json:"appointments"`
	AppointmentsChange    int           `json:"appointments_change"`
	Cancellations         int           `json:"cancellations"`
	CancellationsChange   int           `json:"cancellations_change"`
	AverageDuration       int           `json:"average_duration"`
	AverageDurationChange int           `json:"average_duration_change"`
}

func (s *service) getDashboardStatistics(ctx context.Context, merchantId uuid.UUID, date, currPeriodStart, prevPeriodStart time.Time) (DashboardStatistics, error) {
	query := `
	WITH base AS (
    SELECT
        from_date::date AS day,
        price_then,
        EXTRACT(EPOCH FROM (to_date - from_date)) / 60 AS duration,
        cancelled_by_user_on is not null as cancelled_by_user,
        (cancelled_by_user_on is not null or cancelled_by_merchant_on is not null) as cancelled
    FROM "Appointment"
    WHERE merchant_id = $1
	),
	current AS (
		SELECT
			day,
			SUM(price_then) AS revenue,
			COUNT(*) FILTER (WHERE NOT cancelled) AS appointments,
			COUNT(*) FILTER (WHERE cancelled_by_user) AS cancellations,
			AVG(duration) FILTER (WHERE NOT cancelled) AS avg_duration
		FROM base
		WHERE day >= $2 AND day < $3
		GROUP BY day
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
			COALESCE(SUM(price_then), 0) AS revenue_sum,
			COUNT(*) FILTER (WHERE NOT cancelled) AS appointments,
			COUNT(*) FILTER (WHERE cancelled_by_user) AS cancellations,
			CAST(AVG(duration) FILTER (WHERE NOT cancelled) AS INTEGER) AS average_duration
		FROM base
		WHERE day >= $4 AND day < $5
	)
	SELECT
		ct.revenue_sum, p.revenue_sum,
		ct.appointments, p.appointments,
		ct.cancellations, p.cancellations,
		COALESCE(ct.average_duration, 0), COALESCE(p.average_duration, 0)
	FROM current_totals ct, previous p;
	`

	var stats DashboardStatistics

	var (
		currRevenue, prevRevenue             int
		currAppointments, prevAppointments   int
		currCancellations, prevCancellations int
		currAvgDuration, prevAvgDuration     int
	)

	err := s.db.QueryRow(ctx, query, merchantId, currPeriodStart, date, prevPeriodStart, currPeriodStart).Scan(
		&currRevenue, &prevRevenue,
		&currAppointments, &prevAppointments,
		&currCancellations, &prevCancellations,
		&currAvgDuration, &prevAvgDuration,
	)
	if err != nil {
		return DashboardStatistics{}, err
	}

	stats.RevenueSum = currRevenue
	stats.Appointments = currAppointments
	stats.Cancellations = currCancellations
	stats.AverageDuration = currAvgDuration

	stats.RevenueChange = utils.CalculatePercentChange(prevRevenue, currRevenue)
	stats.AppointmentsChange = utils.CalculatePercentChange(prevAppointments, currAppointments)
	stats.CancellationsChange = utils.CalculatePercentChange(prevCancellations, currCancellations)
	stats.AverageDurationChange = utils.CalculatePercentChange(prevAvgDuration, currAvgDuration)

	query2 := `
	SELECT
		EXTRACT(DAY FROM d.day)::int AS day,
		COALESCE(SUM(a.price_then), 0)::int AS revenue
	FROM generate_series($2::date, $3::date, interval '1 day') AS d(day)
	LEFT JOIN "Appointment" a ON date(a.from_date) = d.day
		AND a.cancelled_by_user_on IS NULL
		AND a.cancelled_by_merchant_on IS NULL
		AND a.merchant_id = $1
	GROUP BY d.day
	ORDER BY d.day;
	`

	rows, err := s.db.Query(ctx, query2, merchantId, currPeriodStart, date)
	if err != nil {
		return DashboardStatistics{}, err
	}
	// nolint:errcheck
	defer rows.Close()

	for rows.Next() {
		var day int
		var revenue int
		if err := rows.Scan(&day, &revenue); err != nil {
			return stats, err
		}
		stats.Revenue = append(stats.Revenue, RevenueStat{
			Revenue: revenue,
			Day:     day,
		})
	}

	return stats, nil
}
