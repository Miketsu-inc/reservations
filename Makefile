include .env

MAKEFLAGS += --no-print-directory

run:
	@make -j 6 vite-jabulani vite-tango air db kv caddy

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
	@npx email dev --dir "backend/emails/templates"

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

connect-river:
	@DATABASE_URL="postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_DATABASE}?sslmode=disable&search_path=${DB_SCHEMA}" PORT="2020" ~/go/bin/riverui

lint:
	@npm run lint
	@golangci-lint run

test:
	@go test -v ./backend/...
	@npm run test