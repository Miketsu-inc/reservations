package domain

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/miketsu-inc/reservations/backend/internal/types"
	"github.com/miketsu-inc/reservations/backend/pkg/db"
)

type MerchantRepository interface {
	WithTx(tx db.DBTX) MerchantRepository

	// Insert a new Merchant to the database, creates the default preferences and an owner employee
	NewMerchant(ctx context.Context, userId uuid.UUID, merchant Merchant) error
	DeleteMerchant(ctx context.Context, employeeId int, merchantId uuid.UUID) error

	ChangeMerchantNameAndURL(ctx context.Context, merchantId uuid.UUID, name string, urlName string) error
	UpdateMerchantFields(ctx context.Context, merchantId uuid.UUID, merchantFields MerchantSettingFields) error

	IsMerchantUrlUnique(ctx context.Context, urlName string) (bool, error)
	GetMerchantIdByUrlName(ctx context.Context, urlName string) (uuid.UUID, error)
	GetMerchantUrlName(ctx context.Context, merchantId uuid.UUID) (string, error)
	GetMerchantTimezone(ctx context.Context, merchantId uuid.UUID) (*time.Location, error)
	GetMerchantCurrency(ctx context.Context, merchantId uuid.UUID) (string, error)
	GetMerchantSubscriptionTier(ctx context.Context, merchantId uuid.UUID) (types.SubTier, error)
	GetAllMerchantInfo(ctx context.Context, merchantId uuid.UUID) (MerchantInfo, error)
	GetMerchantSettingsInfo(ctx context.Context, merchantId uuid.UUID) (MerchantSettingsInfo, error)
	GetBookingSettingsByMerchantAndService(ctx context.Context, merchantId uuid.UUID, serviceId int) (MerchantBookingSettings, error)

	GetDashboardStats(ctx context.Context, merchantId uuid.UUID, startDate time.Time, endDate time.Time, prevStartDate time.Time) (DashboardStatistics, error)
	GetRevenueStats(ctx context.Context, merchantId uuid.UUID, startDate time.Time, endDate time.Time) ([]RevenueStat, error)

	NewBusinessHours(ctx context.Context, merchantId uuid.UUID, businessHours map[int][]TimeSlot) error
	DeleteOutdatedBusinessHours(ctx context.Context, merchantId uuid.UUID, businessHours map[int][]TimeSlot) error
	GetBusinessHours(ctx context.Context, merchantId uuid.UUID) (map[int][]TimeSlot, error)
	GetBusinessHoursForDay(ctx context.Context, merchantId uuid.UUID, day int) ([]TimeSlot, error)
	// Get business hours for merchant including only the first start and last ending time
	GetNormalizedBusinessHours(ctx context.Context, merchantId uuid.UUID) (map[int]TimeSlot, error)

	NewLocation(ctx context.Context, location Location) error
	GetLocation(ctx context.Context, locationId int, merchantId uuid.UUID) (Location, error)

	NewPreferences(ctx context.Context, merchantId uuid.UUID) error
	UpdatePreferences(ctx context.Context, merchantId uuid.UUID, preferences PreferenceData) error
	GetPreferences(ctx context.Context, merchantId uuid.UUID) (PreferenceData, error)
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
	Introduction     string `json:"introduction"`
	Announcement     string `json:"announcement"`
	AboutUs          string `json:"about_us"`
	ParkingInfo      string `json:"parking_info"`
	PaymentInfo      string `json:"payment_info"`
	CancelDeadline   int    `json:"cancel_deadline"`
	BookingWindowMin int    `json:"booking_window_min"`
	BookingWindowMax int    `json:"booking_window_max"`
	BufferTime       int    `json:"buffer_time"`
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
