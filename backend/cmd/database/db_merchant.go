package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
}

func (s *service) NewMerchant(ctx context.Context, merchant Merchant) error {
	query := `
	insert into "Merchant" (ID, name, url_name, owner_id, contact_email, introduction, announcement, about_us, parking_info, payment_info)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.ExecContext(ctx, query, merchant.Id, merchant.Name, merchant.UrlName, merchant.OwnerId, merchant.ContactEmail,
		merchant.Introduction, merchant.Announcement, merchant.AboutUs, merchant.ParkingInfo, merchant.PaymentInfo)
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
	err := s.db.QueryRowContext(ctx, query, UrlName).Scan(&merchantId)
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
	err := s.db.QueryRowContext(ctx, query, ownerId).Scan(&merchantId)
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
	err := s.db.QueryRowContext(ctx, query, merchantId).Scan(&merchant.Id, &merchant.Name, &merchant.UrlName, &merchant.OwnerId, &merchant.ContactEmail,
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

	LocationId int    `json:"location_id"`
	Country    string `json:"country"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Address    string `json:"address"`

	Services []PublicService `json:"services"`

	BusinessHours map[int][]TimeSlots `json:"business_hours"`
}

// this should and will be refactored
func (s *service) GetAllMerchantInfo(ctx context.Context, merchantId uuid.UUID) (MerchantInfo, error) {
	query := `
	select m.name, m.url_name, m.contact_email, m.introduction, m.announcement, m.about_us, m.parking_info, m.payment_info,
		l.id as location_id, l.country, l.city, l.postal_code, l.address from "Merchant" m
	inner join "Location" l on m.id = l.merchant_id
	where m.id = $1
	`

	var mi MerchantInfo
	err := s.db.QueryRowContext(ctx, query, merchantId).Scan(&mi.Name, &mi.UrlName, &mi.ContactEmail, &mi.Introduction, &mi.Announcement,
		&mi.AboutUs, &mi.ParkingInfo, &mi.PaymentInfo, &mi.LocationId, &mi.Country, &mi.City, &mi.PostalCode, &mi.Address)
	if err != nil {
		return MerchantInfo{}, err
	}

	mi.Services, err = s.GetServicesByMerchantId(ctx, merchantId)
	if err != nil {
		return MerchantInfo{}, err
	}

	query = `
	select day_of_week, start_time, end_time from "BusinessHours"
	where merchant_id = $1
	order by day_of_week, start_time;`

	rows, err := s.db.QueryContext(ctx, query, merchantId)
	if err != nil {
		return MerchantInfo{}, fmt.Errorf("failed to get business hours for merchant: %v", err)
	}

	defer rows.Close()

	mi.BusinessHours = make(map[int][]TimeSlots)

	for day := 0; day <= 6; day++ {
		mi.BusinessHours[day] = []TimeSlots{}
	}

	for rows.Next() {
		var dayOfWeek int
		var ts TimeSlots

		err := rows.Scan(&dayOfWeek, &ts.StartTime, &ts.EndTime)
		if err != nil {
			return MerchantInfo{}, err
		}

		mi.BusinessHours[dayOfWeek] = append(mi.BusinessHours[dayOfWeek], ts)
	}

	return mi, nil
}

func (s *service) IsMerchantUrlUnique(ctx context.Context, merchantUrl string) error {
	query := `
	select 1 from "Merchant"
	where url_name = $1
	`

	var url string
	err := s.db.QueryRowContext(ctx, query, merchantUrl).Scan(&url)
	if !errors.Is(err, sql.ErrNoRows) {
		if err != nil {
			return err
		}

		return fmt.Errorf("this merchant url is already used: %s", merchantUrl)
	}

	return nil
}

type PublicCustomer struct {
	Customer
	IsBlacklisted  bool `json:"is_blacklisted"`
	TimesBooked    int  `json:"times_booked"`
	TimesCancelled int  `json:"times_cancelled"`
}

func (s *service) GetCustomersByMerchantId(ctx context.Context, merchantId uuid.UUID) ([]PublicCustomer, error) {
	query := `
	select u.id, u.first_name, u.last_name, u.email, u.phone_number, u.is_dummy, b.user_id is not null as is_blacklisted,
		count(a.id) as times_booked, 0 as times_cancelled
	from "User" u
	left join "Appointment" a on u.id = a.user_id and a.merchant_id = $1
	left join "Blacklist" b on u.id = b.user_id and b.merchant_id = $2
	where u.is_dummy = true or a.id is not null
	group by u.id, b.user_id;
	`

	rows, err := s.db.QueryContext(ctx, query, merchantId, merchantId)
	if err != nil {
		return []PublicCustomer{}, err
	}
	defer rows.Close()

	var customers []PublicCustomer
	for rows.Next() {
		var cust PublicCustomer
		if err := rows.Scan(&cust.Id, &cust.FirstName, &cust.LastName, &cust.Email, &cust.PhoneNumber, &cust.IsDummy, &cust.IsBlacklisted,
			&cust.TimesBooked, &cust.TimesCancelled); err != nil {
			return []PublicCustomer{}, err
		}
		customers = append(customers, cust)
	}

	// if customers array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(customers) == 0 {
		customers = []PublicCustomer{}
	}

	return customers, nil
}

type TimeSlots struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type MerchantSettingsInfo struct {
	Name          string              `json:"merchant_name"`
	ContactEmail  string              `json:"contact_email"`
	Introduction  string              `json:"introduction"`
	Announcement  string              `json:"announcement"`
	AboutUs       string              `json:"about_us"`
	ParkingInfo   string              `json:"parking_info"`
	PaymentInfo   string              `json:"payment_info"`
	BusinessHours map[int][]TimeSlots `json:"business_hours"`

	LocationId int    `json:"location_id"`
	Country    string `json:"country"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	Address    string `json:"address"`
}

func (s *service) GetMerchantSettingsInfo(ctx context.Context, merchantId uuid.UUID) (MerchantSettingsInfo, error) {

	var msi MerchantSettingsInfo

	merchantQuery := `
	select m.name, m.contact_email, m.introduction, m.announcement,
		   m.about_us, m.parking_info, m.payment_info,
	       l.id as location_id, l.country, l.city, l.postal_code, l.address
	from "Merchant" m inner join "Location" l on m.id = l.merchant_id
	where m.id = $1;`

	err := s.db.QueryRowContext(ctx, merchantQuery, merchantId).Scan(&msi.Name, &msi.ContactEmail, &msi.Introduction, &msi.Announcement,
		&msi.AboutUs, &msi.ParkingInfo, &msi.PaymentInfo, &msi.LocationId, &msi.Country, &msi.City, &msi.PostalCode, &msi.Address)
	if err != nil {
		return MerchantSettingsInfo{}, err
	}

	businessHoursQuery := `
	select day_of_week, start_time, end_time from "BusinessHours"
	where merchant_id = $1
	order by day_of_week, start_time;`

	rows, err := s.db.QueryContext(ctx, businessHoursQuery, merchantId)
	if err != nil {
		return MerchantSettingsInfo{}, fmt.Errorf("failed to get business hours for merchant: %v", err)
	}

	defer rows.Close()

	msi.BusinessHours = make(map[int][]TimeSlots)

	for day := 0; day <= 6; day++ {
		msi.BusinessHours[day] = []TimeSlots{}
	}

	for rows.Next() {
		var dayOfWeek int
		var ts TimeSlots

		err := rows.Scan(&dayOfWeek, &ts.StartTime, &ts.EndTime)
		if err != nil {
			return MerchantSettingsInfo{}, err
		}
		msi.BusinessHours[dayOfWeek] = append(msi.BusinessHours[dayOfWeek], ts)
	}

	return msi, nil
}

func (s *service) UpdateMerchantFieldsById(ctx context.Context, merchantId uuid.UUID, introduction, announcement, aboutUs, paymentInfo, parkingInfo string, businessHours map[int][]TimeSlots) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer tx.Rollback()

	merchantQuery := `
	update "Merchant"
	set introduction = $2, announcement = $3, about_us = $4, payment_info = $5, parking_info = $6
	where id = $1;`

	_, err = tx.ExecContext(ctx, merchantQuery, merchantId, introduction, announcement, aboutUs, paymentInfo, parkingInfo)
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

	_, err = tx.ExecContext(ctx, deleteQuery, merchantId, utils.IntSliceToPgArray(days), utils.TimeStringToPgArray(starts), utils.TimeStringToPgArray(ends))
	if err != nil {
		return fmt.Errorf("failed to delete outdated business hours for merchant: %v", err)
	}

	// 2. Insert new rows (avoiding duplicates)
	insertQuery := `
    insert into "BusinessHours" (merchant_id, day_of_week, start_time, end_time)
    select $1, unnest($2::int[]), unnest($3::time[]), unnest($4::time[])
    on conflict (merchant_id, day_of_week, start_time, end_time) do nothing;`

	_, err = tx.ExecContext(ctx, insertQuery, merchantId, utils.IntSliceToPgArray(days), utils.TimeStringToPgArray(starts), utils.TimeStringToPgArray(ends))
	if err != nil {
		return fmt.Errorf("failed to insert business hours for merchant: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *service) GetBusinessHoursByDay(ctx context.Context, merchantId uuid.UUID, dayOfWeek int) ([]TimeSlots, error) {
	query := `
	select start_time, end_time from "BusinessHours"
	where merchant_id = $1 and day_of_week = $2
	order by start_time`

	rows, err := s.db.QueryContext(ctx, query, merchantId, dayOfWeek)
	if err != nil {
		return []TimeSlots{}, err
	}

	defer rows.Close()

	bHours := []TimeSlots{}

	for rows.Next() {
		var ts TimeSlots

		err := rows.Scan(&ts.StartTime, &ts.EndTime)
		if err != nil {
			return []TimeSlots{}, err
		}
		bHours = append(bHours, ts)
	}

	return bHours, nil
}

func (s *service) GetNormalizedBusinessHours(ctx context.Context, merchantId uuid.UUID) (map[int]TimeSlots, error) {
	query := `
	select day_of_week, min(start_time) as start_time,
	max(end_time) as end_time from "BusinessHours"
	where merchant_id = $1
	group by day_of_week
	order by day_of_week;`

	rows, err := s.db.QueryContext(ctx, query, merchantId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make(map[int]TimeSlots)
	for rows.Next() {
		var day int
		var startTime, endTime string

		err := rows.Scan(&day, &startTime, &endTime)
		if err != nil {
			return nil, err
		}

		result[day] = TimeSlots{
			StartTime: startTime,
			EndTime:   endTime,
		}

	}
	return result, nil
}
