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
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

// Service represents a service that interacts with a database.
type PostgreSQL interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	Close()

	// -- Appointment --

	// Insert a new Appointment to the database.
	NewAppointment(context.Context, Appointment) error
	// Get all Aappintments assigned to a Merchant.
	GetAppointmentsByMerchant(context.Context, uuid.UUID, string, string) ([]AppointmentDetails, error)
	// Get all available times for reservations
	GetReservedTimes(context.Context, uuid.UUID, int, time.Time) ([]AppointmentTime, error)
	// Update appointment fields
	UpdateAppointmentData(context.Context, uuid.UUID, int, string, string, string) error
	// Transfer dummy user appointments
	TransferDummyAppointments(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
	// Cancel appointment by merchant
	CancelAppointmentByMerchant(context.Context, uuid.UUID, int, string) error

	// -- User --

	// Insert a new User to the database.
	NewUser(context.Context, User) error
	// Get a User by user id.
	GetUserById(context.Context, uuid.UUID) (User, error)
	// Get a User's password and id by the User's email.
	// Used for comparing password hashes on login and setting jwt cookies.
	GetUserPasswordAndIDByUserEmail(context.Context, string) (uuid.UUID, string, error)
	// Check if an email exists in the database.
	IsEmailUnique(context.Context, string) error
	// Check if a phone number exists in the database.
	IsPhoneNumberUnique(context.Context, string) error
	// Increment User's refresh version, logging out the User.
	IncrementUserJwtRefreshVersion(context.Context, uuid.UUID) error
	// Get User's refresh version
	GetUserJwtRefreshVersion(context.Context, uuid.UUID) (int, error)
	// Check if the user is blacklisted
	IsUserBlacklisted(context.Context, uuid.UUID, uuid.UUID) error

	// -- Merchant --

	// Insert a new Merchant to the database
	NewMerchant(context.Context, Merchant) error
	// Get a Merchant's id by the Merchant's url name
	GetMerchantIdByUrlName(context.Context, string) (uuid.UUID, error)
	// Get a Merchant's owner id by the merchantId
	GetMerchantIdByOwnerId(context.Context, uuid.UUID) (uuid.UUID, error)
	// Get a Merchant by it's user id
	GetMerchantById(context.Context, uuid.UUID) (Merchant, error)
	// Get all publicly available merchant info that will be displayed
	GetAllMerchantInfo(context.Context, uuid.UUID) (MerchantInfo, error)
	// Check if a merchant url exists in the database
	IsMerchantUrlUnique(context.Context, string) error
	// Get all necessary information for merchant settings page
	GetMerchantSettingsInfo(context.Context, uuid.UUID) (MerchantSettingsInfo, error)
	// Update the field used in the reservation page
	UpdateMerchantFieldsById(context.Context, uuid.UUID, string, string, string, string, string, map[int][]TimeSlot) error
	// Get business hours for a merchant by a given day
	GetBusinessHoursByDay(context.Context, uuid.UUID, int) ([]TimeSlot, error)
	// Get business hours for merchant including only the first start aand last ending time
	GetNormalizedBusinessHours(context.Context, uuid.UUID) (map[int]TimeSlot, error)
	// Get the merchant's timezone by it's id
	GetMerchantTimezoneById(context.Context, uuid.UUID) (string, error)

	// -- Location --

	// Insert a new Location to the database
	NewLocation(context.Context, Location) error
	// Get location by it's id
	GetLocationById(context.Context, int) (Location, error)

	// -- Serivce --

	// Insert a new service to the database
	NewService(context.Context, Service) error
	// Get a Service by it's id
	GetServiceById(context.Context, int) (Service, error)
	// Get all services for a merchant by it's id
	GetServicesByMerchantId(context.Context, uuid.UUID) ([]PublicService, error)
	// Delete a Service by it's id
	DeleteServiceById(context.Context, uuid.UUID, int) error
	// Update a Service by it's id
	UpdateServiceById(context.Context, Service) error

	// -- Customer --

	// Get all customers for a mechant by it's id
	GetCustomersByMerchantId(context.Context, uuid.UUID) ([]PublicCustomer, error)
	// Insert a new customer to the database
	NewCustomer(context.Context, uuid.UUID, Customer) error
	// Delete customer by it's id
	DeleteCustomerById(context.Context, uuid.UUID, uuid.UUID) error
	// Update customer by it's id
	UpdateCustomerById(context.Context, uuid.UUID, Customer) error
	// Add customer to the blacklist
	AddCustomerToBlacklist(context.Context, uuid.UUID, uuid.UUID) error
	// Remove customer from the blacklist
	RemoveCustomerFromBlacklist(context.Context, uuid.UUID, uuid.UUID) error

	// -- Preferences --

	// Create default preferences for merchant
	CreatePreferences(context.Context, uuid.UUID) error
	// Get all preferences for a merchant by it's id
	GetPreferencesByMerchantId(context.Context, uuid.UUID) (PreferenceData, error)
	// Update preferences for a merchant
	UpdatePreferences(context.Context, uuid.UUID, PreferenceData) error

	// -- Products --

	// Insert a new product into the database
	NewProduct(context.Context, Product) error
	// Get all products for a merchant by it's id
	GetProductsByMerchant(context.Context, uuid.UUID) ([]PublicProduct, error)
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
