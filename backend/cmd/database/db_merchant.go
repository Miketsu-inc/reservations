package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Merchant struct {
	Id           uuid.UUID        `json:"ID"`
	Name         string           `json:"name"`
	UrlName      string           `json:"url_name"`
	OwnerId      uuid.UUID        `json:"owner_id"`
	ContactEmail string           `json:"contact_email"`
	Introduction string           `json:"introduction"`
	Announcement string           `json:"announcement"`
	AboutUs      string           `json:"about_us"`
	ParkingInfo  string           `json:"parking_info"`
	PaymentInfo  string           `json:"payment_info"`
	Settings     MerchantSettings `json:"settings"`
}

type MerchantSettings struct {
	Test bool `json:"test"`
}

func (ms MerchantSettings) Value() (driver.Value, error) {
	return json.Marshal(ms)
}

func (ms *MerchantSettings) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("column value is not of type []byte")
	}

	return json.Unmarshal(b, &ms)
}

func (s *service) NewMerchant(ctx context.Context, merchant Merchant) error {
	query := `
	insert into "Merchant" (ID, name, url_name, owner_id, contact_email, introduction, announcement, about_us, parking_info, payment_info, settings)
	values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.db.ExecContext(ctx, query, merchant.Id, merchant.Name, merchant.UrlName, merchant.OwnerId, merchant.ContactEmail,
		merchant.Introduction, merchant.Announcement, merchant.AboutUs, merchant.ParkingInfo, merchant.PaymentInfo, merchant.Settings)
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
		&merchant.Introduction, &merchant.Announcement, &merchant.AboutUs, &merchant.ParkingInfo, &merchant.PaymentInfo, &merchant.Settings)
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

	query = `
	select id, name, description, color, duration, price, cost from "Service"
	where merchant_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, merchantId)
	if err != nil {
		return MerchantInfo{}, err
	}
	defer rows.Close()

	var services []PublicService
	for rows.Next() {
		var serv PublicService
		if err := rows.Scan(&serv.Id, &serv.Name, &serv.Description, &serv.Color, &serv.Duration, &serv.Price, &serv.Cost); err != nil {
			return MerchantInfo{}, err
		}

		services = append(services, serv)
	}

	// if services array is empty the encoded json field will be null
	// unless an empty slice is supplied to it
	if len(services) == 0 {
		mi.Services = []PublicService{}
	} else {
		mi.Services = services
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
	Id             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	IsDummy        bool      `json:"is_dummy"`
	TimesBooked    int       `json:"times_booked"`
	TimesCancelled int       `json:"times_cancelled"`
}

func (s *service) GetCustomersByMerchantId(ctx context.Context, merchantId uuid.UUID) ([]PublicCustomer, error) {
	query := `
	select u.id, u.first_name, u.last_name, u.is_dummy, count(a.id) as times_booked, 0 as times_cancelled
	from "User" u
	left join "Appointment" a on u.id = a.user_id and a.merchant_id = $1
	where u.is_dummy = true or a.id is not null
	group by u.id;
	`

	rows, err := s.db.QueryContext(ctx, query, merchantId)
	if err != nil {
		return []PublicCustomer{}, err
	}
	defer rows.Close()

	var customers []PublicCustomer
	for rows.Next() {
		var cust PublicCustomer
		if err := rows.Scan(&cust.Id, &cust.FirstName, &cust.LastName, &cust.IsDummy, &cust.TimesBooked, &cust.TimesCancelled); err != nil {
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
