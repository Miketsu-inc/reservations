include .env

run:
	@make -j 4 tailwindcss vite air db

build:
	@npx @tailwindcss/cli -i ./frontend/src/assets/input.css -o ./frontend/src/assets/output.css --minify
	@npm run build
	@go build -o backend/bin/reservations backend/cmd/main.go

vite:
	@npm run dev

air:
	@air

tailwindcss:
	@npx @tailwindcss/cli -i ./frontend/src/assets/input.css -o ./frontend/src/assets/output.css --watch

db:
	@docker start postgresdb

create-db:
	@docker run --name postgresdb -p ${DB_PORT}:${DB_PORT} -d -e POSTGRES_PASSWORD=${DB_PASSWORD} -e POSTGRES_USER=${DB_USERNAME} -e POSTGRES_DB=${DB_DATABASE} -v pgdata:/var/lib/postgresql/data postgres

connect-db:
	containerID=$(shell docker ps -q -f ancestor=postgres); \
	docker exec -it $$containerID psql -U ${DB_USERNAME} ${DB_DATABASE}

lint:
	@npm run lint
	@golangci-lint run

test:
	@go test ./backend/...
	@npm run test