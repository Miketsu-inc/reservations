package database

import (
	"context"

	"github.com/google/uuid"
)

type Preferences struct {
	Id                 int       `json:"ID"`
	MerchantId         uuid.UUID `json:"merchant _id"`
	FirstDayOfWeek     string    `json:"first_day_of_week"`
	TimeFormat         string    `json:"time_format"`
	CalendarView       string    `json:"calendar_view"`
	CalendarViewMobile string    `json:"calendar_view_mobile"`
}

func (s *service) CreatePreferences(ctx context.Context, merchantId uuid.UUID) error {

	query := `
	insert into "Preferences" (merchant_id) values ($1)`

	_, err := s.db.ExecContext(ctx, query, merchantId)
	if err != nil {
		return err
	}
	return nil
}

type PreferenceData struct {
	FirstDayOfWeek     string `json:"first_day_of_week"`
	TimeFormat         string `json:"time_format"`
	CalendarView       string `json:"calendar_view"`
	CalendarViewMobile string `json:"calendar_view_mobile"`
}

func (s *service) GetPreferencesByMerchantId(ctx context.Context, merchantId uuid.UUID) (PreferenceData, error) {

	query := `
	select first_day_of_week, time_format, calendar_view, calendar_view_mobile from "Preferences"
	where merchant_id = $1`

	var p PreferenceData
	err := s.db.QueryRowContext(ctx, query, merchantId).Scan(&p.FirstDayOfWeek, &p.TimeFormat, &p.CalendarView, &p.CalendarViewMobile)
	if err != nil {
		return PreferenceData{}, err
	}

	return p, nil
}

func (s *service) UpdatePreferences(ctx context.Context, merchantId uuid.UUID, p PreferenceData) error {
	query := `
	update "Preferences"
	set first_day_of_week = $2, time_format = $3, calendar_view = $4, calendar_view_mobile = $5
	where merchant_id = $1;`

	_, err := s.db.ExecContext(ctx, query, merchantId, p.FirstDayOfWeek, p.TimeFormat, p.CalendarView, p.CalendarViewMobile)
	if err != nil {
		return err
	}

	return nil
}
