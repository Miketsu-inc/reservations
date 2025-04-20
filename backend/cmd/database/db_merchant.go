package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
