# Reservations

## Setting up the development environment

Required binaries

- [Nodejs](https://nodejs.org/en)
- [Go](https://go.dev/)
- [Docker](https://www.docker.com/)
- [Air](https://github.com/air-verse/air)
- [Make](https://www.gnu.org/software/make/)
- [Caddy](https://caddyserver.com/)
- [Golangci-lint](https://golangci-lint.run/) (for local linting)

Installing dependencies

```
npm install
go mod tidy
```

You need the following environment variables in a .env file at the root of the project to successfully run the application.

```
PORT
APP_ENV

DB_HOST
DB_PORT
DB_DATABASE
DB_USERNAME
DB_PASSWORD
DB_SCHEMA

KV_PORT

JWT_ACCESS_SECRET
JWT_ACCESS_EXP_MIN
JWT_REFRESH_SECRET
JWT_REFRESH_EXP_MIN

RESEND_API_TEST
ENABLE_EMAILS

VITE_MAPBOX_TOKEN

OAUTH_STATE_SECRET
GOOGLE_OAUTH_CLIENT_ID
GOOGLE_OAUTH_CLIENT_SECRET
FACEBOOK_OAUTH_CLIENT_ID
FACEBOOK_OAUTH_CLIENT_SECRET

TANGO_URL
JABULANI_URL
DOMAIN

VITE_TANGO_URL=$TANGO_URL
VITE_JABULANI_URL=$JABULANI_URL

R2_ENDPOINT
R2_API_ACCESS_KEY_ID
R2_API_SECRET_ACCESS_KEY
R2_UPLOADS_BUCKET
R2_MEDIA_BUCKET
MEDIA_PUBLIC_URL

IMGPROXY_PORT
IMGPROXY_URL
IMGPROXY_KEY
IMGPROXY_SALT
```

Images are stored using [Cloudflare R2](https://developers.cloudflare.com/r2/) buckets. You need the following setup.

- `R2_UPLOADS_BUCKET` should be private, have a lifecycle rule deleting objects after 1 day and a CORS rule allowing `PUT` requests with the `Content-Type` header from the `TANGO_URL` and `JABULANI_URL` origins.
- `R2_MEDIA_BUCKET` should be public, `MEDIA_PUBLIC_URL` being the public url
- An API token with Object Read & Write permission for these two buckets.

Uploaded images are processed by [imgproxy](https://docs.imgproxy.net), create a client using the `create-imgproxy` command.

```
make create-imgproxy
```

Create a Redis client in docker using the `create-kv` command.

```
make create-kv
```

Create a PostgreSQL database in docker using the `create-db` command.

```
make create-db
```

Install [riverqueue](https://riverqueue.com/docs) (make sure it's the correct version) and run the migrations.

```
go install github.com/riverqueue/river/cmd/river
make migrate-river
```

Use the `make run` command to run the application in development mode.

```
make run
```

To have https use the following caddy command.

```
caddy trust
```

## Building from source

After setting up the development environment you can build the application using the make build command.

```
make build
```
