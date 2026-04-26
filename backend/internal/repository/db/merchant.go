package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bojanz/currency"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/miketsu-inc/reservations/backend/internal/domain"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/internal/utils"
	"github.com/miketsu-inc/reservations/backend/pkg/currencyx"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type merchantRepository struct {
	db db.DBTX
}

func NewMerchantRepository(db db.DBTX) domain.MerchantRepository {
	return &merchantRepository{db: db}
}

func (r *merchantRepository) WithTx(tx db.DBTX) domain.MerchantRepository {
	return &merchantRepository{db: tx}
}

func (r *merchantRepository) NewMerchant(ctx context.Context, userId uuid.UUID, merchant domain.Merchant) error {
	newMerchantQuery := `
	insert into "Merchant" (ID, name, url_name, contact_email, introduction, announcement, about_us, parking_info,
		payment_info, timezone, currency_code, subscription_tier)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(ctx, newMerchantQuery, merchant.Id, merchant.Name, merchant.UrlName, merchant.ContactEmail,
		merchant.Introduction, merchant.Announcement, merchant.AboutUs, merchant.ParkingInfo, merchant.PaymentInfo,
		merchant.Timezone, merchant.CurrencyCode, merchant.SubscriptionTier)
	if err != nil {
		return err
	}

	return nil
}

func (r *merchantRepository) DeleteMerchant(ctx context.Context, employeeId int, merchantId uuid.UUID) error {
	query := `
	delete from "Merchant" m
	using "Employee" e
	where e.user_id = $1 and e.role = 'owner' and m.id = $2
	`

	_, err := r.db.Exec(ctx, query, employeeId, merchantId)
	if err != nil {
		return err
	}

	return nil
}

func (r *merchantRepository) ChangeMerchantNameAndURL(ctx context.Context, merchantId uuid.UUID, MerchantName, urlName string) error {
	query := `
	update "Merchant"
	set name = $2, url_name = $3
	where id = $1`

	_, err := r.db.Exec(ctx, query, merchantId, MerchantName, urlName)
	if err != nil {
		return err
	}

	return nil
}

func (r *merchantRepository) UpdateMerchantFields(ctx context.Context, merchantId uuid.UUID, ms domain.MerchantSettingFields) error {
	query := `
	update "Merchant"
	set introduction = $2, announcement = $3, about_us = $4, payment_info = $5,
	parking_info = $6, cancel_deadline = $7, booking_window_min = $8, booking_window_max = $9, buffer_time = $10, approval_policy = $11
	where id = $1;`

	_, err := r.db.Exec(ctx, query, merchantId, ms.Introduction, ms.Announcement, ms.AboutUs, ms.PaymentInfo, ms.ParkingInfo,
		ms.CancelDeadline, ms.BookingWindowMin, ms.BookingWindowMax, ms.BufferTime, ms.ApprovalPolicy)
	if err != nil {
		return err
	}

	return nil
}

func (r *merchantRepository) IsMerchantUrlUnique(ctx context.Context, merchantUrl string) (bool, error) {
	query := `
	select 1 from "Merchant"
	where url_name = $1
	`

	var exists int
	err := r.db.QueryRow(ctx, query, merchantUrl).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}

	if err != nil {
		return false, err
	}

	return false, nil
}

func (r *merchantRepository) GetMerchantIdByUrlName(ctx context.Context, UrlName string) (uuid.UUID, error) {
	query := `
	select id from "Merchant"
	where url_name = $1
	`

	var merchantId uuid.UUID
	err := r.db.QueryRow(ctx, query, UrlName).Scan(&merchantId)
	if err != nil {
		return uuid.Nil, err
	}

	return merchantId, nil
}

func (r *merchantRepository) GetMerchantUrlName(ctx context.Context, merchantId uuid.UUID) (string, error) {
	query := `
	select url_name
	from "Merchant"
	where id = $1
	`

	var urlName string
	err := r.db.QueryRow(ctx, query, merchantId).Scan(&urlName)
	if err != nil {
		return "", err
	}

	return urlName, nil
}

func (r *merchantRepository) GetMerchantTimezone(ctx context.Context, merchantId uuid.UUID) (*time.Location, error) {
	query := `
	select timezone from "Merchant"
	where id = $1
	`

	var timzone string
	err := r.db.QueryRow(ctx, query, merchantId).Scan(&timzone)
	if err != nil {
		return nil, err
	}

	tz, err := time.LoadLocation(timzone)
	if err != nil {
		return nil, fmt.Errorf("error while parsing merchant's timezone: %s", err.Error())
	}

	return tz, nil
}

func (r *merchantRepository) GetMerchantCurrency(ctx context.Context, merchantId uuid.UUID) (string, error) {
	query := `
	select currency_code from "Merchant" where id = $1
	`

	var curr string
	err := r.db.QueryRow(ctx, query, merchantId).Scan(&curr)
	if err != nil {
		return "", err
	}

	return curr, nil
}

func (r *merchantRepository) GetMerchantSubscriptionTier(ctx context.Context, merchantId uuid.UUID) (types.SubTier, error) {
	query := `
	select subscription_tier from "Merchant"
	where id = $1
	`

	var tier types.SubTier
	err := r.db.QueryRow(ctx, query, merchantId).Scan(&tier)
	if err != nil {
		return types.SubTier{}, err
	}

	return tier, nil
}

// TODO: this should be refactored ideally to one query
func (r *merchantRepository) GetAllMerchantInfo(ctx context.Context, merchantId uuid.UUID) (domain.MerchantInfo, error) {
	query := `
	select m.name, m.url_name, m.contact_email, m.introduction, m.announcement, m.about_us, m.parking_info, m.payment_info, m.timezone,
	l.id as location_id, l.country, l.city, l.postal_code, l.address, l.formatted_location, l.geo_point from "Merchant" m
	inner join "Location" l on m.id = l.merchant_id
	where m.id = $1
	`

	var mi domain.MerchantInfo
	err := r.db.QueryRow(ctx, query, merchantId).Scan(&mi.Name, &mi.UrlName, &mi.ContactEmail, &mi.Introduction, &mi.Announcement,
		&mi.AboutUs, &mi.ParkingInfo, &mi.PaymentInfo, &mi.Timezone, &mi.LocationId, &mi.Country, &mi.City, &mi.PostalCode, &mi.Address,
		&mi.FormattedLocation, &mi.GeoPoint)
	if err != nil {
		return domain.MerchantInfo{}, err
	}

	// this is fine until refactored into one query
	catalogRepo := catalogRepository{db: r.db}
	mi.Services, err = catalogRepo.GetServicesForMerchantPage(ctx, merchantId)
	if err != nil {
		return domain.MerchantInfo{}, err
	}

	businnessHours, err := r.GetBusinessHours(ctx, merchantId)
	if err != nil {
		return domain.MerchantInfo{}, fmt.Errorf("failed to get business hours for merchant: %v", err)
	}

	mi.BusinessHours = businnessHours

	return mi, nil
}

func (r *merchantRepository) GetMerchantSettingsInfo(ctx context.Context, merchantId uuid.UUID) (domain.MerchantSettingsInfo, error) {

	var msi domain.MerchantSettingsInfo

	merchantQuery := `
	select m.name, m.contact_email, m.introduction, m.announcement,
		   m.about_us, m.parking_info, m.payment_info, m.cancel_deadline, m.booking_window_min, m.booking_window_max, m.buffer_time, m.approval_policy, m.timezone,
	       l.id as location_id, l.country, l.city, l.postal_code, l.address, l.formatted_location
	from "Merchant" m inner join "Location" l on m.id = l.merchant_id
	where m.id = $1;`

	err := r.db.QueryRow(ctx, merchantQuery, merchantId).Scan(&msi.Name, &msi.ContactEmail, &msi.Introduction, &msi.Announcement,
		&msi.AboutUs, &msi.ParkingInfo, &msi.PaymentInfo, &msi.CancelDeadline, &msi.BookingWindowMin, &msi.BookingWindowMax, &msi.BufferTime, &msi.ApprovalPolicy,
		&msi.Timezone, &msi.LocationId, &msi.Country, &msi.City, &msi.PostalCode, &msi.Address, &msi.FormattedLocation)
	if err != nil {
		return domain.MerchantSettingsInfo{}, err
	}

	businessHours, err := r.GetBusinessHours(ctx, merchantId)
	if err != nil {
		return domain.MerchantSettingsInfo{}, fmt.Errorf("failed to get business hours for merchant: %v", err)
	}

	msi.BusinessHours = businessHours

	return msi, nil
}

func (r *merchantRepository) GetBookingSettingsByMerchantAndService(ctx context.Context, merchantId uuid.UUID, serviceId int) (domain.MerchantBookingSettings, error) {
	query := `
	select coalesce(s.buffer_time, m.buffer_time) as buffer_time,
	       coalesce(s.booking_window_max, m.booking_window_max) as booking_window_max,
		   coalesce(s.booking_window_min, m.booking_window_min) as booking_window_min,
		   coalesce(s.approval_policy, m.approval_policy) as approval_policy
	from "Merchant" m
	join "Service" s on s.merchant_id = $1
	where m.id = $1 and s.id = $2`

	var mbs domain.MerchantBookingSettings
	err := r.db.QueryRow(ctx, query, merchantId, serviceId).Scan(&mbs.BufferTime, &mbs.BookingWindowMax, &mbs.BookingWindowMin, &mbs.ApprovalPolicy)
	if err != nil {
		return domain.MerchantBookingSettings{}, err
	}

	return mbs, nil
}

func (r *merchantRepository) GetDashboardStats(ctx context.Context, merchantId uuid.UUID, startDate, endDate, prevStartDate time.Time) (domain.DashboardStatistics, error) {
	query := `
	WITH participant as (
		SELECT
			booking_id,
			BOOL_OR(status = 'cancelled') as cancelled_by_user
		FROM "BookingParticipant"
		GROUP BY booking_id
	),
	base AS (
		SELECT
			to_date,
			(total_price).number as price,
			(total_price).currency as currency,
			EXTRACT(EPOCH FROM (to_date - from_date)) / 60 AS duration,
			COALESCE(bp.cancelled_by_user, FALSE) as cancelled_by_user,
			(b.status in ('cancelled')) as cancelled
		FROM "Booking" b
		left join participant bp on bp.booking_id = b.id
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

	var stats domain.DashboardStatistics

	var (
		currRevenue, prevRevenue             int
		currBookings, prevBookings           int
		currCancellations, prevCancellations int
		currAvgDuration, prevAvgDuration     int
		curr                                 string
	)

	err := r.db.QueryRow(ctx, query, merchantId, startDate, endDate, prevStartDate, startDate.AddDate(0, 0, 1)).Scan(
		&currRevenue, &prevRevenue,
		&currBookings, &prevBookings,
		&currCancellations, &prevCancellations,
		&currAvgDuration, &prevAvgDuration,
		&curr,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return domain.DashboardStatistics{}, err
		}
	}

	var formattedRevenue string

	// if no rows are returned
	if curr != "" {
		amount, err := currency.NewAmount(strconv.Itoa(currRevenue), curr)
		if err != nil {
			return domain.DashboardStatistics{}, fmt.Errorf("new amount creation failed: %v", err)
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

	return stats, nil
}

// TODO: in the future only completed bookings should count towards revenue
func (r *merchantRepository) GetRevenueStats(ctx context.Context, merchantId uuid.UUID, startDate, endDate time.Time) ([]domain.RevenueStat, error) {
	query := `
	SELECT
		DATE(bookings.from_date) AS day,
		COALESCE(SUM(bookings.price), 0) AS value
	FROM (
		select b.from_date, (b.total_price).number as price
		from "Booking" b
		where b.merchant_id = $1 AND b.from_date >= $2 AND b.from_date < $3 and b.status not in ('cancelled')
		order by b.id
	) as bookings
	GROUP BY day
	ORDER BY day
	`

	rows, _ := r.db.Query(ctx, query, merchantId, startDate, endDate)
	revenue, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.RevenueStat])
	if err != nil {
		return []domain.RevenueStat{}, err
	}

	return revenue, nil
}

func (r *merchantRepository) NewBusinessHours(ctx context.Context, merchantId uuid.UUID, businessHours domain.BusinessHours) error {
	query := `
	insert into "BusinessHours" (merchant_id, day_of_week, start_time, end_time)
    select $1, unnest($2::int[]), unnest($3::time[]), unnest($4::time[])
    on conflict (merchant_id, day_of_week, start_time, end_time) do nothing
	`

	days := make([]int, 0)
	startTimes := make([]time.Time, 0)
	endTimes := make([]time.Time, 0)

	for day, timeRanges := range businessHours {
		for _, ts := range timeRanges {
			if !ts.StartTime.IsZero() && !ts.EndTime.IsZero() {
				days = append(days, day)
				startTimes = append(startTimes, ts.StartTime)
				endTimes = append(endTimes, ts.EndTime)
			}
		}
	}

	_, err := r.db.Exec(ctx, query, merchantId, days, startTimes, endTimes)
	if err != nil {
		return err
	}

	return nil
}

func (r *merchantRepository) DeleteOutdatedBusinessHours(ctx context.Context, merchantId uuid.UUID, businessHours domain.BusinessHours) error {
	query := `
	delete from "BusinessHours"
    where merchant_id = $1
    and (day_of_week, start_time, end_time) not in (
        select unnest($2::int[]), unnest($3::time[]), unnest($4::time[])
    )
	`

	days := make([]int, 0)
	startTimes := make([]time.Time, 0)
	endTimes := make([]time.Time, 0)

	for day, timeRanges := range businessHours {
		for _, ts := range timeRanges {
			if !ts.StartTime.IsZero() && !ts.EndTime.IsZero() {
				days = append(days, day)
				startTimes = append(startTimes, ts.StartTime)
				endTimes = append(endTimes, ts.EndTime)
			}
		}
	}

	_, err := r.db.Exec(ctx, query, merchantId, days, startTimes, endTimes)
	if err != nil {
		return err
	}

	return nil
}

func (r *merchantRepository) GetBusinessHours(ctx context.Context, merchantId uuid.UUID) (domain.BusinessHours, error) {
	query := `
	select day_of_week, start_time, end_time from "BusinessHours"
	where merchant_id = $1
	order by day_of_week, start_time;
	`

	businessHours := make(domain.BusinessHours)
	for day := 0; day <= 6; day++ {
		businessHours[day] = []domain.TimeSlot{}
	}

	var dayOfWeek int
	var start, end time.Time
	rows, _ := r.db.Query(ctx, query, merchantId)
	_, err := pgx.ForEachRow(rows, []any{&dayOfWeek, &start, &end}, func() error {
		businessHours[dayOfWeek] = append(businessHours[dayOfWeek], domain.TimeSlot{
			StartTime: start,
			EndTime:   end,
		})

		return nil
	})
	if err != nil {
		return domain.BusinessHours{}, err
	}

	return businessHours, nil
}

func (r *merchantRepository) GetBusinessHoursForDay(ctx context.Context, merchantId uuid.UUID, dayOfWeek int) ([]domain.TimeSlot, error) {
	query := `
	select start_time, end_time from "BusinessHours"
	where merchant_id = $1 and day_of_week = $2
	order by start_time`

	rows, _ := r.db.Query(ctx, query, merchantId, dayOfWeek)
	bHours, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.TimeSlot, error) {
		var ts domain.TimeSlot
		err := row.Scan(&ts.StartTime, &ts.EndTime)

		return ts, err
	})
	if err != nil {
		return []domain.TimeSlot{}, err
	}

	return bHours, nil
}

func (r *merchantRepository) GetNormalizedBusinessHours(ctx context.Context, merchantId uuid.UUID) (domain.BusinessHours, error) {
	query := `
	select day_of_week, min(start_time) as start_time,
	max(end_time) as end_time from "BusinessHours"
	where merchant_id = $1
	group by day_of_week
	order by day_of_week;`

	rows, _ := r.db.Query(ctx, query, merchantId)

	var day int
	var startTime, endTime time.Time

	result := make(domain.BusinessHours)
	_, err := pgx.ForEachRow(rows, []any{&day, &startTime, &endTime}, func() error {
		result[day] = []domain.TimeSlot{{
			StartTime: startTime,
			EndTime:   endTime,
		}}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *merchantRepository) NewLocation(ctx context.Context, location domain.Location) error {
	query := `
	insert into "Location" (merchant_id, country, city, postal_code, address, geo_point, place_id, formatted_location, is_primary, is_active)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(ctx, query, location.MerchantId, location.Country, location.City, location.PostalCode, location.Address, location.GeoPoint,
		location.PlaceId, location.FormattedLocation, location.IsPrimary, location.IsActive)
	if err != nil {
		return err
	}

	return nil
}

func (r *merchantRepository) GetLocation(ctx context.Context, locationId int, merchantId uuid.UUID) (domain.Location, error) {
	query := `
	select * from "Location"
	where id = $1 and merchant_id = $2
	`

	var location domain.Location
	err := r.db.QueryRow(ctx, query, locationId, merchantId).Scan(&location.Id, &location.MerchantId, &location.Country, &location.City, &location.PostalCode,
		&location.Address, &location.GeoPoint, &location.PlaceId, &location.FormattedLocation, &location.IsPrimary, &location.IsActive)
	if err != nil {
		return domain.Location{}, err
	}

	return location, nil
}

func (r *merchantRepository) NewPreferences(ctx context.Context, merchantId uuid.UUID) error {
	query := `
	insert into "Preferences" (merchant_id) values ($1)
	`

	_, err := r.db.Exec(ctx, query, merchantId)
	if err != nil {
		return err
	}

	return err
}

func (r *merchantRepository) UpdatePreferences(ctx context.Context, merchantId uuid.UUID, p domain.PreferenceData) error {
	query := `
	update "Preferences"
	set first_day_of_week = $2, time_format = $3, calendar_view = $4, calendar_view_mobile = $5, start_hour = $6, end_hour = $7, time_frequency = $8
	where merchant_id = $1;`

	_, err := r.db.Exec(ctx, query, merchantId, p.FirstDayOfWeek, p.TimeFormat, p.CalendarView, p.CalendarViewMobile, p.StartHour, p.EndHour, p.TimeFrequency)
	if err != nil {
		return err
	}

	return nil
}

func (r *merchantRepository) GetPreferences(ctx context.Context, merchantId uuid.UUID) (domain.PreferenceData, error) {

	query := `
	select first_day_of_week, time_format, calendar_view, calendar_view_mobile, start_hour, end_hour, time_frequency from "Preferences"
	where merchant_id = $1`

	var p domain.PreferenceData
	err := r.db.QueryRow(ctx, query, merchantId).Scan(&p.FirstDayOfWeek, &p.TimeFormat, &p.CalendarView, &p.CalendarViewMobile, &p.StartHour, &p.EndHour, &p.TimeFrequency)
	if err != nil {
		return domain.PreferenceData{}, err
	}

	return p, nil
}
