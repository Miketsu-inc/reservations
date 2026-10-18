package config

import (
	"os"
	"strconv"
	"sync"

	_ "github.com/joho/godotenv/autoload"
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

	KV_PORT string

	JWT_ACCESS_SECRET   string
	JWT_ACCESS_EXP_MIN  int
	JWT_REFRESH_SECRET  string
	JWT_REFRESH_EXP_MIN int

	RESEND_API_TEST string
	ENABLE_EMAILS   bool

	OAUTH_STATE_SECRET           string
	GOOGLE_OAUTH_CLIENT_ID       string
	GOOGLE_OAUTH_CLIENT_SECRET   string
	FACEBOOK_OAUTH_CLIENT_ID     string
	FACEBOOK_OAUTH_CLIENT_SECRET string

	TANGO_URL    string
	JABULANI_URL string
	DOMAIN       string

	R2_ACCOUNT_ID        string
	R2_ACCESS_KEY_ID     string
	R2_SECRET_ACCESS_KEY string
	R2_UPLOADS_BUCKET    string
	R2_MEDIA_BUCKET      string
	MEDIA_PUBLIC_URL     string
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
		kv_port := os.Getenv("KV_PORT")
		jwt_access_secret := os.Getenv("JWT_ACCESS_SECRET")
		jwt_access_exp_min, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_EXP_MIN"))
		jwt_refresh_secret := os.Getenv("JWT_REFRESH_SECRET")
		jwt_refresh_exp_min, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_EXP_MIN"))
		resend_api_test := os.Getenv("RESEND_API_TEST")
		enable_emails, _ := strconv.ParseBool(os.Getenv("ENABLE_EMAILS"))
		oauth_state_secret := os.Getenv("OAUTH_STATE_SECRET")
		google_oauth_client_id := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
		google_oauth_client_secret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
		facebook_oauth_client_id := os.Getenv("FACEBOOK_OAUTH_CLIENT_ID")
		facebook_oauth_client_secret := os.Getenv("FACEBOOK_OAUTH_CLIENT_SECRET")
		tango_url := os.Getenv("TANGO_URL")
		jabulani_url := os.Getenv("JABULANI_URL")
		domain := os.Getenv("DOMAIN")
		r2_account_id := os.Getenv("R2_ACCOUNT_ID")
		r2_access_key_id := os.Getenv("R2_ACCESS_KEY_ID")
		r2_secret_access_key := os.Getenv("R2_SECRET_ACCESS_KEY")
		r2_uploads_bucket := os.Getenv("R2_UPLOADS_BUCKET")
		r2_media_bucket := os.Getenv("R2_MEDIA_BUCKET")
		media_public_url := os.Getenv("MEDIA_PUBLIC_URL")

		instance = &Config{
			PORT:                         port,
			APP_ENV:                      app_env,
			DB_HOST:                      db_host,
			DB_PORT:                      db_port,
			DB_DATABASE:                  db_database,
			DB_USERNAME:                  db_username,
			DB_PASSWORD:                  db_password,
			DB_SCHEMA:                    db_schema,
			KV_PORT:                      kv_port,
			JWT_ACCESS_SECRET:            jwt_access_secret,
			JWT_ACCESS_EXP_MIN:           jwt_access_exp_min,
			JWT_REFRESH_SECRET:           jwt_refresh_secret,
			JWT_REFRESH_EXP_MIN:          jwt_refresh_exp_min,
			RESEND_API_TEST:              resend_api_test,
			ENABLE_EMAILS:                enable_emails,
			OAUTH_STATE_SECRET:           oauth_state_secret,
			GOOGLE_OAUTH_CLIENT_ID:       google_oauth_client_id,
			GOOGLE_OAUTH_CLIENT_SECRET:   google_oauth_client_secret,
			FACEBOOK_OAUTH_CLIENT_ID:     facebook_oauth_client_id,
			FACEBOOK_OAUTH_CLIENT_SECRET: facebook_oauth_client_secret,
			TANGO_URL:                    tango_url,
			JABULANI_URL:                 jabulani_url,
			DOMAIN:                       domain,
			R2_ACCOUNT_ID:                r2_account_id,
			R2_ACCESS_KEY_ID:             r2_access_key_id,
			R2_SECRET_ACCESS_KEY:         r2_secret_access_key,
			R2_UPLOADS_BUCKET:            r2_uploads_bucket,
			R2_MEDIA_BUCKET:              r2_media_bucket,
			MEDIA_PUBLIC_URL:             media_public_url,
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
	assert.True(c.KV_PORT != "", "KV_PORT environment variable could not be found")
	assert.True(c.JWT_ACCESS_SECRET != "", "JWT_ACCESS_SECRET environment variable could not be found")
	assert.True(c.JWT_ACCESS_EXP_MIN != 0, "JWT_ACCESS_EXP_MIN environment variable could not be found")
	assert.True(c.JWT_REFRESH_SECRET != "", "JWT_REFRESH_SECRET environment variable could not be found")
	assert.True(c.JWT_REFRESH_EXP_MIN != 0, "JWT_REFRESH_EXP_MIN environment variable could not be found")
	assert.True(c.RESEND_API_TEST != "", "RESEND_API_TEST environment variable could not be found")
	assert.NotNil(c.ENABLE_EMAILS, "ENABLE_EMAILS environment variable could not be found")
	assert.True(c.OAUTH_STATE_SECRET != "", "OAUTH_STATE_SECRET environment variable could not be found")
	assert.True(c.GOOGLE_OAUTH_CLIENT_ID != "", "GOOGLE_OAUTH_CLIENT_ID environment variable could not be found")
	assert.True(c.GOOGLE_OAUTH_CLIENT_SECRET != "", "GOOGLE_OAUTH_CLIENT_SECRET environment variable could not be found")
	assert.True(c.FACEBOOK_OAUTH_CLIENT_ID != "", "FACEBOOK_OAUTH_CLIENT_ID environment variable could not be found")
	assert.True(c.FACEBOOK_OAUTH_CLIENT_SECRET != "", "FACEBOOK_OAUTH_CLIENT_SECRET environment variable could not be found")
	assert.True(c.TANGO_URL != "", "TANGO_URL environment variable could not be found")
	assert.True(c.JABULANI_URL != "", "JABULANI_URL environment variable could not be found")
	assert.True(c.DOMAIN != "", "DOMAIN environment variable could not be found")
	assert.True(c.R2_ACCOUNT_ID != "", "R2_ACCOUNT_ID environment variable could not be found")
	assert.True(c.R2_ACCESS_KEY_ID != "", "R2_ACCESS_KEY_ID environment variable could not be found")
	assert.True(c.R2_SECRET_ACCESS_KEY != "", "R2_SECRET_ACCESS_KEY environment variable could not be found")
	assert.True(c.R2_UPLOADS_BUCKET != "", "R2_UPLOADS_BUCKET environment variable could not be found")
	assert.True(c.R2_MEDIA_BUCKET != "", "R2_MEDIA_BUCKET environment variable could not be found")
	assert.True(c.MEDIA_PUBLIC_URL != "", "MEDIA_PUBLIC_URL environment variable could not be found")
}

func (c *Config) IsProd() bool {
	return c.APP_ENV == "prod"
}
