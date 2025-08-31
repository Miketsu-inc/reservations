package database

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Preferences struct {
	Id                 int        `json:"ID"`
	MerchantId         uuid.UUID  `json:"merchant _id"`
	FirstDayOfWeek     string     `json:"first_day_of_week"`
	TimeFormat         string     `json:"time_format"`
	CalendarView       string     `json:"calendar_view"`
	CalendarViewMobile string     `json:"calendar_view_mobile"`
	StartHour          TimeString `json:"start_hour"`
	EndHour            TimeString `json:"end_hour"`
	TimeFrequency      TimeString `json:"time_frequency"`
}

type TimeString string

func (ts TimeString) MarshalJSON() ([]byte, error) {
	timeStr := string(ts)
	if strings.Contains(timeStr, ".") {
		if parsed, err := time.Parse("15:04:05.000000", timeStr); err == nil {
			timeStr = parsed.Format("15:04:05")
		}
	}
	return json.Marshal(timeStr)
}

func (ts *TimeString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*ts = TimeString(s)

	return nil
}

func (s *service) CreatePreferences(ctx context.Context, merchantId uuid.UUID) error {
	query := `
	insert into "Preferences" (merchant_id) values ($1)`

	_, err := s.db.Exec(ctx, query, merchantId)
	if err != nil {
		return err
	}
	return nil
}

type PreferenceData struct {
	FirstDayOfWeek     string     `json:"first_day_of_week"`
	TimeFormat         string     `json:"time_format"`
	CalendarView       string     `json:"calendar_view"`
	CalendarViewMobile string     `json:"calendar_view_mobile"`
	StartHour          TimeString `json:"start_hour"`
	EndHour            TimeString `json:"end_hour"`
	TimeFrequency      TimeString `json:"time_frequency"`
}

func (s *service) GetPreferencesByMerchantId(ctx context.Context, merchantId uuid.UUID) (PreferenceData, error) {

	query := `
	select first_day_of_week, time_format, calendar_view, calendar_view_mobile, start_hour, end_hour, time_frequency from "Preferences"
	where merchant_id = $1`

	var p PreferenceData
	err := s.db.QueryRow(ctx, query, merchantId).Scan(&p.FirstDayOfWeek, &p.TimeFormat, &p.CalendarView, &p.CalendarViewMobile, &p.StartHour, &p.EndHour, &p.TimeFrequency)
	if err != nil {
		return PreferenceData{}, err
	}

	return p, nil
}

func (s *service) UpdatePreferences(ctx context.Context, merchantId uuid.UUID, p PreferenceData) error {
	query := `
	update "Preferences"
	set first_day_of_week = $2, time_format = $3, calendar_view = $4, calendar_view_mobile = $5, start_hour = $6, end_hour = $7, time_frequency = $8
	where merchant_id = $1;`

	_, err := s.db.Exec(ctx, query, merchantId, p.FirstDayOfWeek, p.TimeFormat, p.CalendarView, p.CalendarViewMobile, p.StartHour, p.EndHour, p.TimeFrequency)
	if err != nil {
		return err
	}

	return nil
}
