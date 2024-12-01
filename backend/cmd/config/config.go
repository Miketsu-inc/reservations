package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/miketsu-inc/reservations/backend/pkg/assert"
)

type Config struct {
	PORT    string
	APP_ENV string

	DB_HOST     string
	DB_PORT     string
	DB_DATABASE string
	DB_USERNAME string
	DB_PASSWORD string
	DB_SCHEMA   string

	JWT_ACCESS_SECRET   string
	JWT_ACCESS_EXP_MIN  int
	JWT_REFRESH_SECRET  string
	JWT_REFRESH_EXP_MIN int

	EMAIL_ADDRESS  string
	EMAIL_PASSWORD string
	SMTP_HOST      string
	SMTP_PORT      string
}

var instance *Config
var once sync.Once

func LoadEnvVars() *Config {
	once.Do(func() {
		port := os.Getenv("PORT")
		assert.True(port != "", "PORT environment variable could not be found")

		app_env := os.Getenv("APP_ENV")
		assert.True(app_env != "", "APP_ENV environment variable could not be found")

		db_host := os.Getenv("DB_HOST")
		assert.True(db_host != "", "DB_HOST environment variable could not be found")

		db_port := os.Getenv("DB_PORT")
		assert.True(db_port != "", "DB_PORT environment variable could not be found")

		db_database := os.Getenv("DB_DATABASE")
		assert.True(db_database != "", "DB_DATABASE environment variable could not be found")

		db_username := os.Getenv("DB_USERNAME")
		assert.True(db_username != "", "DB_USERNAME environment variable could not be found")

		db_password := os.Getenv("DB_PASSWORD")
		assert.True(db_password != "", "DB_PASSWORD environment variable could not be found")

		db_schema := os.Getenv("DB_SCHEMA")
		assert.True(db_schema != "", "DB_SCHEMA environment variable could not be found")

		jwt_access_secret := os.Getenv("JWT_ACCESS_SECRET")
		assert.True(jwt_access_secret != "", "JWT_ACCESS_SECRET environment variable could not be found")

		jwt_access_exp_min, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_EXP_MIN"))
		assert.True(jwt_access_exp_min != 0, "JWT_ACCESS_EXP_MIN environment variable could not be found")

		jwt_refresh_secret := os.Getenv("JWT_REFRESH_SECRET")
		assert.True(jwt_refresh_secret != "", "JWT_REFRESH_SECRET environment variable could not be found")

		jwt_refresh_exp_min, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_EXP_MIN"))
		assert.True(jwt_refresh_exp_min != 0, "JWT_REFRESH_EXP_MIN environment variable could not be found")

		email_address := os.Getenv("EMAIL_ADDRESS")
		assert.True(email_address != "", "EMAIL_ADDRESS environment variable could not be found")

		email_password := os.Getenv("EMAIL_PASSWORD")
		assert.True(email_password != "", "EMAIL_PASSWORD environment variable could not be found")

		smtp_host := os.Getenv("SMTP_HOST")
		assert.True(smtp_host != "", "SMTP_HOST environment variable could not be found")

		smtp_port := os.Getenv("SMTP_PORT")
		assert.True(smtp_port != "", "SMTP_PORT environment variable could not be found")

		instance = &Config{
			PORT:                port,
			APP_ENV:             app_env,
			DB_HOST:             db_host,
			DB_PORT:             db_port,
			DB_DATABASE:         db_database,
			DB_USERNAME:         db_username,
			DB_PASSWORD:         db_password,
			DB_SCHEMA:           db_schema,
			JWT_ACCESS_SECRET:   jwt_access_secret,
			JWT_ACCESS_EXP_MIN:  jwt_access_exp_min,
			JWT_REFRESH_SECRET:  jwt_refresh_secret,
			JWT_REFRESH_EXP_MIN: jwt_refresh_exp_min,
			EMAIL_ADDRESS:       email_address,
			EMAIL_PASSWORD:      email_password,
			SMTP_HOST:           smtp_host,
			SMTP_PORT:           smtp_port,
		}
	})
	return instance
}