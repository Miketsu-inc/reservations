package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

// Service represents a service that interacts with a database.
type PostgreSQL interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// -- Appointment --

	// Insert a new Appointment to the database.
	NewAppointment(context.Context, Appointment) error
	// Get all Appointments made by a User.
	GetAppointmentsByUser(context.Context, uuid.UUID) ([]Appointment, error)
	// Get all Aappintments assigned to a Merchant.
	GetAppointmentsByMerchant(context.Context, uuid.UUID, string, string) ([]Appointment, error)

	// -- Auth --

	// Insert a new User to the database.
	NewUser(context.Context, User) error
	// Get a User by user id.
	GetUserById(context.Context, uuid.UUID) (User, error)
	// Get a User's password by the User's email.
	// Used for comparing password hashes on login.
	GetUserPasswordByUserEmail(context.Context, string) (string, error)
	// Check if an email exists in the database
	IsEmailUnique(context.Context, string) bool
	// Check if a phone number exists in the database
	IsPhoneNumberUnique(context.Context, string) bool

	// -- Merchant --

	// Insert a new Merchant to the database
	NewMerchant(context.Context, Merchant) error
	// Get a Merchant's id by the Merchant's url name
	GetMerchantIdByUrlName(context.Context, string) (uuid.UUID, error)
	// Get a Merchant's owner id by the merchantId
	GetMerchantIdByOwnerId(context.Context, uuid.UUID) (uuid.UUID, error)
	// Get a Merchant by it's user id
	GetMerchantById(context.Context, uuid.UUID) (Merchant, error)

	// -- Location --

	// Insert a new Location to the database
	NewLocation(context.Context, Location) error

	// -- Serivce --

	// Insert new services to the database
	NewServices(context.Context, []Service) error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() PostgreSQL {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	assert.Nil(err, "PostgreSQL database could not be openned", err)

	dbInstance = &service{
		db: db,
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
	err := s.db.PingContext(ctx)
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
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
