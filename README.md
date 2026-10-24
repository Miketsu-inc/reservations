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

R2_ENDPOINT
R2_ACCESS_KEY_ID
R2_SECRET_ACCESS_KEY
R2_UPLOADS_BUCKET
R2_MEDIA_BUCKET
MEDIA_PUBLIC_URL

IMGPROXY_PORT
IMGPROXY_URL
IMGPROXY_KEY
IMGPROXY_SALT

VITE_TANGO_URL=$TANGO_URL
VITE_JABULANI_URL=$JABULANI_URL
```

Images are stored in two Cloudflare R2 buckets.

`R2_ENDPOINT` is the S3 API url of the account, `https://<account id>.r2.cloudflarestorage.com` (without a bucket name).

- `R2_UPLOADS_BUCKET` is private, clients upload the original files to it with presigned urls. It needs a lifecycle rule deleting objects after 1 day and a CORS rule allowing `PUT` requests with the `Content-Type` header from the `TANGO_URL` and `JABULANI_URL` origins.
- `R2_MEDIA_BUCKET` is public and contains the processed images. `MEDIA_PUBLIC_URL` is it's public url (the `r2.dev` url or a custom domain).

The API token only needs Object Read & Write permission for these two buckets.

Uploaded images are processed by [imgproxy](https://docs.imgproxy.net), which reads the originals from the uploads bucket. `IMGPROXY_PORT` is the port of the local docker container, `IMGPROXY_URL` is the url the backend reaches imgproxy at (`http://localhost:${IMGPROXY_PORT}` locally). `IMGPROXY_KEY` and `IMGPROXY_SALT` are two random hex strings (e.g. 64 characters each), the backend signs the imgproxy requests with them so only it can use imgproxy. Any random generator works, for example `openssl rand -hex 32`. Create the imgproxy container in docker using the create-imgproxy command, it has to be recreated if any of these environment variables change.

```
make create-imgproxy
```

Create a PostgreSQL database in docker using the create-db command.

```
make create-db
```

Use the make run command to run the application in development mode.

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
