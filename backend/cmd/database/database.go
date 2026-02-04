package database

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/miketsu-inc/reservations/backend/cmd/config"
	"github.com/miketsu-inc/reservations/backend/cmd/types"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
	"golang.org/x/text/language"
)

// Service represents a service that interacts with a database.
type PostgreSQL interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	Close()

	// -- Booking --

	// Insert a new Booking made by the customer.
	NewBookingByCustomer(context.Context, NewCustomerBooking) (int, error)
	// Insert a new Booking made by the merchant.
	NewBookingByMerchant(context.Context, NewMerchantBooking) (int, error)
	// Get all calendar events assigned to a Merchant ina given time period.
	GetCalendarEventsByMerchant(context.Context, uuid.UUID, string, string) (CalendarEvents, error)
	// Get all available times for reservations
	GetReservedTimes(context.Context, uuid.UUID, int, time.Time) ([]BookingTime, error)
	// Get all reserved times for reservations in a given period
	GetReservedTimesForPeriod(context.Context, uuid.UUID, int, time.Time, time.Time) ([]BookingTime, error)
	// Update booking fields
	UpdateBookingData(context.Context, uuid.UUID, int, string, time.Duration) error
	// Transfer dummy user bookings
	TransferDummyBookings(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	// Cancel booking by merchant
	CancelBookingByMerchant(context.Context, uuid.UUID, int, string) error
	// Update the email id for a booking
	UpdateEmailIdForBooking(context.Context, int, string) error
	// Get booking info for email sending
	GetBookingDataForEmail(context.Context, int) (BookingEmailData, error)
	// Get public booking info for user
	GetPublicBookingInfo(context.Context, int) (PublicBookingInfo, error)
	// Cancel booking by customer
	CancelBookingByCustomer(context.Context, uuid.UUID, int) (uuid.UUID, error)
	// Cretate recurring booking instances in batch
	BatchCreateRecurringBookings(context.Context, NewRecurringBookings) (int, error)
	// Get all existing booking dates for a booking series in a time range
	GetExistingOccurrenceDates(context.Context, int, time.Time, time.Time) ([]time.Time, error)
	// Create a new booking series and related tables
	NewBookingSeries(context.Context, NewBookingSeries) (CompleteBookingSeries, error)
	// Get all group bookings from a merchant which is not full in a given period
	GetAvailableGroupBookingsForPeriod(context.Context, uuid.UUID, int, int, time.Time, time.Time) ([]BookingTime, error)

	// -- User --

	// Insert a new User to the database.
	NewUser(context.Context, User) error
	// Get a User by user id.
	GetUserById(context.Context, uuid.UUID) (User, error)
	// Get a User's password and id by the User's email.
	// Used for comparing password hashes on login and setting jwt cookies.
	GetUserPasswordAndIDByUserEmail(context.Context, string) (uuid.UUID, *string, error)
	// Check if an email exists in the database.
	IsEmailUnique(context.Context, string) error
	// Check if a phone number exists in the database.
	IsPhoneNumberUnique(context.Context, string) error
	// Increment User's refresh version, logging out the User.
	IncrementUserJwtRefreshVersion(context.Context, uuid.UUID) error
	// Get User's refresh version
	GetUserJwtRefreshVersion(context.Context, uuid.UUID) (int, error)
	// Get a user's preferred language
	GetUserPreferredLanguage(context.Context, uuid.UUID) (*language.Tag, error)
	// Get all employees associated with a user
	GetEmployeesByUser(context.Context, uuid.UUID) ([]EmployeeAuthInfo, error)
	// Find Oauth user by provider type and id
	FindOauthUser(context.Context, types.AuthProviderType, string) (uuid.UUID, error)

	// -- Merchant --

	// Insert a new Merchant to the database, creates the default preferences and an owner employee
	NewMerchant(context.Context, uuid.UUID, Merchant) error
	// Get a Merchant's id by the Merchant's url name
	GetMerchantIdByUrlName(context.Context, string) (uuid.UUID, error)
	// Get all publicly available merchant info that will be displayed
	GetAllMerchantInfo(context.Context, uuid.UUID) (MerchantInfo, error)
	// Check if a merchant url exists in the database
	IsMerchantUrlUnique(context.Context, string) error
	// Get all necessary information for merchant settings page
	GetMerchantSettingsInfo(context.Context, uuid.UUID) (MerchantSettingsInfo, error)
	// Update the field used in the reservation page
	UpdateMerchantFieldsById(context.Context, uuid.UUID, MerchantSettingFields) error
	// Update a merchant's businessHours
	UpdateBusinessHours(context.Context, uuid.UUID, map[int][]TimeSlot) error
	// Get all businessHours for a merchant
	GetBusinessHours(context.Context, uuid.UUID) (map[int][]TimeSlot, error)
	// Get business hours for a merchant by a given day
	GetBusinessHoursByDay(context.Context, uuid.UUID, int) ([]TimeSlot, error)
	// Get business hours for merchant including only the first start and last ending time
	GetNormalizedBusinessHours(context.Context, uuid.UUID) (map[int]TimeSlot, error)
	// Get the merchant's timezone by it's id
	GetMerchantTimezoneById(context.Context, uuid.UUID) (*time.Location, error)
	// Get the dashboard data by the merchant's id for a period of days
	GetDashboardData(context.Context, uuid.UUID, time.Time, int) (DashboardData, error)
	// Get the merchant's currency
	GetMerchantCurrency(context.Context, uuid.UUID) (string, error)
	// Get the merchant's subscription tier
	GetMerchantSubscriptionTier(context.Context, uuid.UUID) (types.SubTier, error)
	// Get necessary booking settings by a merchnat's id
	GetBookingSettingsByMerchantAndService(context.Context, uuid.UUID, int) (MerchantBookingSettings, error)
	// Delete Merchant from the database by the employee and merchant id
	DeleteMerchant(context.Context, int, uuid.UUID) error
	// Change merchant's name and url name
	ChangeMerchantNameAndURL(context.Context, uuid.UUID, string, string) error
	// Get the merchant's url name by id
	GetMerchantUrlNameById(context.Context, uuid.UUID) (string, error)
	// Create new blocked time for one or multiple employees
	NewBlockedTime(context.Context, uuid.UUID, []int, string, time.Time, time.Time, bool, *int) ([]int, error)
	// Delete bloced time for an employee by id
	DeleteBlockedTime(context.Context, int, uuid.UUID, int) error
	// Update blocked time for an employee
	UpdateBlockedTime(context.Context, BlockedTime) error
	// Get blocked times for available calculation
	GetBlockedTimes(context.Context, uuid.UUID, time.Time, time.Time) ([]BlockedTimes, error)
	// Get all blocked time type by a merchant's id
	GetAllBlockedTimeTypes(context.Context, uuid.UUID) ([]BlockedTimeType, error)
	// Create a blocked time type for a merchant
	NewBlockedTimeType(context.Context, uuid.UUID, BlockedTimeType) error
	// Update blocked time type for a merchant by id
	UpdateBlockedTimeType(context.Context, uuid.UUID, BlockedTimeType) error
	// Delete blocked time type for a merchant by id
	DeleteBlockedTimeType(context.Context, uuid.UUID, int) error
	// Get employees by merchant
	GetEmployeesForCalendarByMerchant(context.Context, uuid.UUID) ([]EmployeeForCalendar, error)
	// Get employees for merchant
	GetEmployeesByMerchant(context.Context, uuid.UUID) ([]PublicEmployee, error)
	// Get employee by id
	GetEmployeeById(context.Context, uuid.UUID, int) (PublicEmployee, error)
	// Create new employee
	NewEmployee(context.Context, uuid.UUID, PublicEmployee) error
	// Update employee by id
	UpdateEmployeeById(context.Context, uuid.UUID, PublicEmployee) error
	// Delete employee by id
	DeleteEmployeeById(context.Context, uuid.UUID, int) error
	// New external calendar
	NewExternalCalendar(context.Context, ExternalCalendar) (int, error)
	// Update external calendar sync token
	UpdateExternalCalendarSyncToken(context.Context, int, string) error
	// Bulk insert rows for initial calendar sync (BlockedTime, ExternalCalendarEvent)
	BulkInitialSyncExternalCalendarEvents(context.Context, []BlockedTime, []int, []ExternalCalendarEvent) error
	// Bulk insert, update, delete rows from incremental calendar sync (BlockedTime, ExternalCalendarEvent)
	BulkIncrementalSyncExternalCalendarEvents(context.Context, []BlockedTime, []BlockedTime, []int, []int, []ExternalCalendarEvent,
		[]ExternalCalendarEvent, []ExternalEventBlockedTimeLink) error
	// Get all external calendar events by external event ids
	GetExternalCalendarEventsByIds(context.Context, int, []string) ([]ExternalCalendarEvent, error)
	// Delete all external calendar related data (BlockedTime, ExternalCalendarEvent) and reset sync state
	// should be called for 410 GONE response before full initial sync
	ResetExternalCalendar(context.Context, int) error
	// Get the external calendar for an employee by their id
	GetExternalCalendarByEmployeeId(context.Context, int) (ExternalCalendar, error)
	// Update access, refresh tokens and their expiry for an external calendar
	UpdateExternalCalendarAuthTokens(context.Context, int, string, string, time.Time) error

	// -- Location --

	// Insert a new Location to the database
	NewLocation(context.Context, Location) error
	// Get location by it's id
	GetLocationById(context.Context, int, uuid.UUID) (Location, error)

	// -- Service --

	// Insert a new service to the database
	NewService(context.Context, Service, []ServicePhase, []ConnectedProducts) error
	// Get a Service and it's phases by it's id
	GetServiceWithPhasesById(context.Context, int, uuid.UUID) (PublicServiceWithPhases, error)
	// Get all services for a merchant by it's id
	GetServicesByMerchantId(context.Context, uuid.UUID) ([]ServicesGroupedByCategory, error)
	// Delete a Service by it's id
	DeleteServiceById(context.Context, uuid.UUID, int) error
	// Update a Service and it's phases by it's id
	UpdateServiceWithPhaseseById(context.Context, ServiceWithPhasesAndSettings) error
	// Deactivate a service by it's id
	DeactivateServiceById(context.Context, uuid.UUID, int) error
	// Activate a service by it's id
	ActivateServiceById(context.Context, uuid.UUID, int) error
	// Reorder services
	ReorderServices(context.Context, uuid.UUID, *int, []int) error
	// Insert a new service category
	NewServiceCategory(context.Context, uuid.UUID, ServiceCategory) error
	// Update a service category by it's id
	UpdateServiceCategoryById(context.Context, uuid.UUID, ServiceCategory) error
	// Delete a service category by it's id, making it's services uncategorized
	DeleteServiceCategoryById(context.Context, uuid.UUID, int) error
	// Reorder service categories
	ReorderServiceCategories(context.Context, uuid.UUID, []int) error
	// Get all data related to a service
	GetAllServicePageData(context.Context, int, uuid.UUID) (ServicePageData, error)
	// Get all additional data required for the service page
	GetServicePageFormOptions(context.Context, uuid.UUID) (ServicePageFormOptions, error)
	// Update the connected product for a service by id
	UpdateConnectedProducts(context.Context, int, []ConnectedProducts) error
	// Get all services grouped by category for the merchant page
	GetServicesForMerchantPage(context.Context, uuid.UUID) ([]MerchantPageServicesGroupedByCategory, error)
	// get public details about a service for the merchant page service details section
	GetServiceDetailsForMerchantPage(context.Context, uuid.UUID, int, int) (PublicServiceDetails, error)
	// get simple service info for the merchant page booking summary
	GetMinimalServiceInfo(context.Context, uuid.UUID, int, int) (MinimalServiceInfo, error)
	// Get all services for calendar
	GetServicesForCalendarByMerchant(context.Context, uuid.UUID) ([]ServiceForCalendar, error)
	// Update Group Service and it's phase by id
	UpdateGroupServiceById(context.Context, GroupServiceWithSettings) error
	// Get all data related to a group service
	GetGroupServicePageData(context.Context, uuid.UUID, int) (GroupServicePageData, error)

	// -- Customer --

	// Get all customers for a mechant by it's id
	GetCustomersByMerchantId(context.Context, uuid.UUID, bool) ([]PublicCustomer, error)
	// Insert a new customer to the database
	NewCustomer(context.Context, uuid.UUID, Customer) error
	// Delete customer by it's id
	DeleteCustomerById(context.Context, uuid.UUID, uuid.UUID) error
	// Update customer by it's id
	UpdateCustomerById(context.Context, uuid.UUID, Customer) error
	// Set blacklist status for a customer
	SetBlacklistStatusForCustomer(context.Context, uuid.UUID, uuid.UUID, bool, *string) error
	// Get one customer's info and bookings for a merchant
	GetCustomerStatsByMerchant(context.Context, uuid.UUID, uuid.UUID) (CustomerStatistics, error)
	//Get a User's customer id from it's user id and the merchant's id
	GetCustomerIdByUserIdAndMerchantId(context.Context, uuid.UUID, uuid.UUID) (uuid.UUID, error)
	// Get one customer's info for a merchant
	GetCustomerInfoByMerchant(context.Context, uuid.UUID, uuid.UUID) (CustomerInfo, error)
	// Get a customer's email by id
	GetCustomerEmailById(context.Context, uuid.UUID, uuid.UUID) (string, error)
	// Get all customers for calendar
	GetCustomersForCalendarByMerchant(context.Context, uuid.UUID) ([]CustomerForCalendar, error)

	// -- Preferences --

	// Get all preferences for a merchant by it's id
	GetPreferencesByMerchantId(context.Context, uuid.UUID) (PreferenceData, error)
	// Update preferences for a merchant
	UpdatePreferences(context.Context, uuid.UUID, PreferenceData) error

	// -- Products --

	// Insert a new product into the database
	NewProduct(context.Context, Product) error
	// Get all products for a merchant by it's id
	GetProductsByMerchant(context.Context, uuid.UUID) ([]ProductInfo, error)
	// Delete a Product by it's id
	DeleteProductById(context.Context, uuid.UUID, int) error
	// Updateing properties of product by a it's id
	UpdateProduct(context.Context, Product) error
}

type service struct {
	db *pgxpool.Pool
}

var (
	dbInstance *service
)

func New() PostgreSQL {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	cfg := config.LoadEnvVars()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", cfg.DB_USERNAME, cfg.DB_PASSWORD, cfg.DB_HOST, cfg.DB_PORT, cfg.DB_DATABASE, cfg.DB_SCHEMA)
	dbpool, err := pgxpool.New(context.Background(), connStr)
	assert.Nil(err, "PostgreSQL database could not be openned", err)

	dbInstance = &service{
		db: dbpool,
	}

	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.Ping(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("%s", fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stat()
	stats["open_connections"] = strconv.Itoa(int(dbStats.AcquiredConns()))
	stats["in_use"] = strconv.Itoa(int(dbStats.TotalConns()))
	stats["idle"] = strconv.Itoa(int(dbStats.IdleConns()))
	stats["wait_count"] = strconv.FormatInt(dbStats.AcquireCount(), 10)
	stats["wait_duration"] = dbStats.AcquireDuration().String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleDestroyCount(), 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeDestroyCount(), 10)

	// Evaluate stats to provide a health message
	if dbStats.AcquiredConns() > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.AcquireCount() > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleDestroyCount() > int64(dbStats.AcquiredConns())/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeDestroyCount() > int64(dbStats.AcquiredConns())/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
func (s *service) Close() {
	s.db.Close()
}
