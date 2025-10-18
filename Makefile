include .env

MAKEFLAGS += --no-print-directory

run:
	@make -j 4 tailwindcss vite air db

build:
	@npx @tailwindcss/cli -i ./frontend/src/assets/input.css -o ./frontend/src/assets/output.css --minify
	@npm run build
	@make email-build
	@make go-build

vite:
	@npm run dev

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

tailwindcss:
	@npx @tailwindcss/cli -i ./frontend/src/assets/input.css -o ./frontend/src/assets/output.css --watch

email:
	@npx email dev --dir "backend/emails/templates"

email-build:
	@npx email export --dir "backend/emails/templates" --outDir "backend/emails/out" --pretty

db:
	@docker start postgresdb

create-db:
	@docker run --name postgresdb -p ${DB_PORT}:${DB_PORT} -d -e POSTGRES_PASSWORD=${DB_PASSWORD} -e POSTGRES_USER=${DB_USERNAME} -e POSTGRES_DB=${DB_DATABASE} -v pgdata:/var/lib/postgresql/data postgres

connect-db:
	@docker exec -it postgresdb psql -U ${DB_USERNAME} ${DB_DATABASE}

lint:
	@npm run lint
	@golangci-lint run

test:
	@go test -v ./backend/...
	@npm run test