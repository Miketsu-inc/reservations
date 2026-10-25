include .env

MAKEFLAGS += --no-print-directory

DATABASE_URL = postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE}?sslmode=disable&search_path=${DB_SCHEMA}

run:
	@make -j 7 vite-jabulani vite-tango air db kv imgproxy caddy

build:
	@npm run build-jabulani
	@npm run build-tango
ifeq (${skip-email},)
	@make email-build
endif
	@make go-build

vite-jabulani:
	@npm run dev-jabulani

vite-tango:
	@npm run dev-tango

ifeq ($(OS),Windows_NT)
air:
	@air -build.cmd "go build -o backend/bin/reservations.exe backend/cmd/main.go" -build.bin "backend\bin\reservations.exe"

go-build:
	@go build -tags=prod -o backend/bin/reservations.exe backend/cmd/main.go

else
air:
	@air

go-build:
	@go build -tags=prod -o backend/bin/reservations backend/cmd/main.go

endif

email:
	@npx dotenv -e .env -v EMAIL_PREVIEW=1 -- npx email dev --dir "backend/emails/templates"

email-build:
	@npx dotenv -e .env -- npx email export --dir "backend/emails/templates" --outDir "backend/emails/out" --pretty

caddy:
	@caddy run --config Caddyfile

db:
	@docker start postgresdb

create-db:
	@docker run --name postgresdb -p ${DB_PORT}:${DB_PORT} -d -e POSTGRES_PASSWORD=${DB_PASSWORD} -e POSTGRES_USER=${DB_USERNAME} -e POSTGRES_DB=${DB_DATABASE} -v pgdata:/var/lib/postgresql/data postgis/postgis

connect-db:
	@docker exec -it postgresdb psql -U ${DB_USERNAME} ${DB_DATABASE}

kv:
	@docker start redis

create-kv:
	@docker run --name redis -p ${KV_PORT}:${KV_PORT} -d redis

connect-kv:
	@docker exec -it redis redis-cli

imgproxy:
	@docker start imgproxy

create-imgproxy:
	@docker run --name imgproxy -p ${IMGPROXY_PORT}:8080 -d -e IMGPROXY_KEY=${IMGPROXY_KEY} -e IMGPROXY_SALT=${IMGPROXY_SALT} -e IMGPROXY_USE_S3=true -e IMGPROXY_S3_ENDPOINT=${R2_ENDPOINT} -e IMGPROXY_S3_REGION=auto -e AWS_ACCESS_KEY_ID=${R2_API_ACCESS_KEY_ID} -e AWS_SECRET_ACCESS_KEY=${R2_API_SECRET_ACCESS_KEY} -e IMGPROXY_ALLOWED_SOURCES=s3://${R2_UPLOADS_BUCKET}/ -e IMGPROXY_MAX_SRC_FILE_SIZE=10485760 -e IMGPROXY_MAX_SRC_RESOLUTION=50 ghcr.io/imgproxy/imgproxy:v4

connect-river:
	@DATABASE_URL="${DATABASE_URL}" PORT="2020" ~/go/bin/riverui

migrate-river:
	@~/go/bin/river migrate-up --database-url "${DATABASE_URL}"

lint:
	@npm run lint
	@golangci-lint run

test:
	@go test -v ./backend/...
	@npm run test