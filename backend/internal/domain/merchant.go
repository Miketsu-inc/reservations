package domain

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
)

type MerchantRepository interface {
	// Insert a new Merchant to the database, creates the default preferences and an owner employee
	NewMerchant(context.Context, uuid.UUID, Merchant) error
	DeleteMerchant(context.Context, int, uuid.UUID) error

	GetMerchantIdByUrlName(context.Context, string) (uuid.UUID, error)
	GetMerchantUrlNameById(context.Context, uuid.UUID) (string, error)
	IsMerchantUrlUnique(context.Context, string) error
	ChangeMerchantNameAndURL(context.Context, uuid.UUID, string, string) error

	GetAllMerchantInfo(context.Context, uuid.UUID) (MerchantInfo, error)
	GetMerchantSettingsInfo(context.Context, uuid.UUID) (MerchantSettingsInfo, error)
	GetBookingSettingsByMerchantAndService(context.Context, uuid.UUID, int) (MerchantBookingSettings, error)

	UpdateMerchantFieldsById(context.Context, uuid.UUID, MerchantSettingFields) error

	GetBusinessHours(context.Context, uuid.UUID) (map[int][]TimeSlot, error)
	GetBusinessHoursByDay(context.Context, uuid.UUID, int) ([]TimeSlot, error)
	// Get business hours for merchant including only the first start and last ending time
	GetNormalizedBusinessHours(context.Context, uuid.UUID) (map[int]TimeSlot, error)
	UpdateBusinessHours(context.Context, uuid.UUID, map[int][]TimeSlot) error

	GetMerchantTimezoneById(context.Context, uuid.UUID) (*time.Location, error)
	GetMerchantCurrency(context.Context, uuid.UUID) (string, error)
	GetMerchantSubscriptionTier(context.Context, uuid.UUID) (types.SubTier, error)

	NewLocation(context.Context, Location) error
	GetLocationById(context.Context, int, uuid.UUID) (Location, error)

	GetPreferencesByMerchantId(context.Context, uuid.UUID) (PreferenceData, error)
	UpdatePreferences(context.Context, uuid.UUID, PreferenceData) error

	GetDashboardData(context.Context, uuid.UUID, time.Time, int) (DashboardData, error)
}

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

type TimeSlot struct {
	StartTime string `json:"start_time" db:"start_time"`
	EndTime   string `json:"end_time" db:"end_time"`
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

type MerchantBookingSettings struct {
	BookingWindowMin int `json:"booking_window_min" db:"booking_window_min"`
	BookingWindowMax int `json:"booking_window_max" db:"booking_window_max"`
	BufferTime       int `json:"buffer_time" db:"buffer_time"`
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

type Location struct {
	Id                int            `json:"ID"`
	MerchantId        uuid.UUID      `json:"merchant_id"`
	Country           *string        `json:"country"`
	City              *string        `json:"city"`
	PostalCode        *string        `json:"postal_code"`
	Address           *string        `json:"address"`
	GeoPoint          types.GeoPoint `json:"geo_point"`
	PlaceId           *string        `json:"place_id"`
	FormattedLocation string         `json:"formatted_location"`
	IsPrimary         bool           `json:"is_primary"`
	IsActive          bool           `json:"is_active"`
}

// TODO: surely this is not great
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

func (ts TimeString) String() string {
	timeStr := string(ts)

	if strings.Contains(timeStr, ".") {
		if parsed, err := time.Parse("15:04:05.000000", timeStr); err == nil {
			timeStr = parsed.Format("15:04:05")
		}
	}

	return timeStr
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
