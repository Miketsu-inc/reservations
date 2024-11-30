# Reservations

## Setting up the development environment

> Please note every command in the Makefile was designed to run on a linux system with bash. There can be commands which do not work in Windows or MacOS

Required binaries

- [Nodejs](https://nodejs.org/en)
- [Go](https://go.dev/)
- [Docker](https://www.docker.com/)
- [Air](https://github.com/air-verse/air)
- [Make](https://www.gnu.org/software/make/)

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

JWT_ACCESS_SECRET
JWT_ACCESS_EXP_MIN
JWT_REFRESH_SECRET
JWT_REFRESH_EXP_MIN

EMAIL_ADDRESS
EMAIL_PASSWORD
SMTP_HOST
SMTP_PORT
```

Create a PostgreSQL database in docker using the create-db command.

```
make create-db
```

Use the make run command to run the application in development mode.

```
make run
```

## Building from source

After setting up the development environment you can build the application using the make build command.

```
make build
```
