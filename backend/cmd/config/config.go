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

	RESEND_API_TEST string
	ENABLE_EMAILS   bool

	GOOGLE_OAUTH_CLIENT_ID       string
	GOOGLE_OAUTH_CLIENT_SECRET   string
	FACEBOOK_OAUTH_CLIENT_ID     string
	FACEBOOK_OAUTH_CLIENT_SECRET string
}

var instance *Config
var once sync.Once

func LoadEnvVars() *Config {
	once.Do(func() {
		port := os.Getenv("PORT")
		app_env := os.Getenv("APP_ENV")
		db_host := os.Getenv("DB_HOST")
		db_port := os.Getenv("DB_PORT")
		db_database := os.Getenv("DB_DATABASE")
		db_username := os.Getenv("DB_USERNAME")
		db_password := os.Getenv("DB_PASSWORD")
		db_schema := os.Getenv("DB_SCHEMA")
		jwt_access_secret := os.Getenv("JWT_ACCESS_SECRET")
		jwt_access_exp_min, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_EXP_MIN"))
		jwt_refresh_secret := os.Getenv("JWT_REFRESH_SECRET")
		jwt_refresh_exp_min, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_EXP_MIN"))
		resend_api_test := os.Getenv("RESEND_API_TEST")
		enable_emails, _ := strconv.ParseBool(os.Getenv("ENABLE_EMAILS"))
		google_oauth_client_id := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
		google_oauth_client_secret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
		facebook_oauth_client_id := os.Getenv("FACEBOOK_OAUTH_CLIENT_ID")
		facebook_oauth_client_secret := os.Getenv("FACEBOOK_OAUTH_CLIENT_SECRET")

		instance = &Config{
			PORT:                         port,
			APP_ENV:                      app_env,
			DB_HOST:                      db_host,
			DB_PORT:                      db_port,
			DB_DATABASE:                  db_database,
			DB_USERNAME:                  db_username,
			DB_PASSWORD:                  db_password,
			DB_SCHEMA:                    db_schema,
			JWT_ACCESS_SECRET:            jwt_access_secret,
			JWT_ACCESS_EXP_MIN:           jwt_access_exp_min,
			JWT_REFRESH_SECRET:           jwt_refresh_secret,
			JWT_REFRESH_EXP_MIN:          jwt_refresh_exp_min,
			RESEND_API_TEST:              resend_api_test,
			ENABLE_EMAILS:                enable_emails,
			GOOGLE_OAUTH_CLIENT_ID:       google_oauth_client_id,
			GOOGLE_OAUTH_CLIENT_SECRET:   google_oauth_client_secret,
			FACEBOOK_OAUTH_CLIENT_ID:     facebook_oauth_client_id,
			FACEBOOK_OAUTH_CLIENT_SECRET: facebook_oauth_client_secret,
		}
	})
	return instance
}

func (c *Config) Validate() {
	assert.True(c.PORT != "", "PORT environment variable could not be found")
	assert.True(c.APP_ENV != "", "APP_ENV environment variable could not be found")
	assert.True(c.DB_HOST != "", "DB_HOST environment variable could not be found")
	assert.True(c.DB_PORT != "", "DB_PORT environment variable could not be found")
	assert.True(c.DB_DATABASE != "", "DB_DATABASE environment variable could not be found")
	assert.True(c.DB_USERNAME != "", "DB_USERNAME environment variable could not be found")
	assert.True(c.DB_PASSWORD != "", "DB_PASSWORD environment variable could not be found")
	assert.True(c.DB_SCHEMA != "", "DB_SCHEMA environment variable could not be found")
	assert.True(c.JWT_ACCESS_SECRET != "", "JWT_ACCESS_SECRET environment variable could not be found")
	assert.True(c.JWT_ACCESS_EXP_MIN != 0, "JWT_ACCESS_EXP_MIN environment variable could not be found")
	assert.True(c.JWT_REFRESH_SECRET != "", "JWT_REFRESH_SECRET environment variable could not be found")
	assert.True(c.JWT_REFRESH_EXP_MIN != 0, "JWT_REFRESH_EXP_MIN environment variable could not be found")
	assert.True(c.RESEND_API_TEST != "", "RESEND_API_TEST environment variable could not be found")
	assert.NotNil(c.ENABLE_EMAILS, "ENABLE_EMAILS environment variable could not be found")
	assert.True(c.GOOGLE_OAUTH_CLIENT_ID != "", "GOOGLE_OAUTH_CLIENT_ID environment variable could not be found")
	assert.True(c.GOOGLE_OAUTH_CLIENT_SECRET != "", "GOOGLE_OAUTH_CLIENT_SECRET environment variable could not be found")
	assert.True(c.FACEBOOK_OAUTH_CLIENT_ID != "", "FACEBOOK_OAUTH_CLIENT_ID environment variable could not be found")
	assert.True(c.FACEBOOK_OAUTH_CLIENT_SECRET != "", "FACEBOOK_OAUTH_CLIENT_SECRET environment variable could not be found")
}
