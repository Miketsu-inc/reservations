run:
	@make -j 3 tailwindcss vite air

build:
	@npx tailwindcss -o ./frontend/src/assets/output.css --minify
	@npm run build
	@go build -o backend/bin/reservations backend/cmd/main.go

vite:
	@npm run dev

air:
	@air

tailwindcss:
	@npx tailwindcss -i ./frontend/src/assets/input.css -o ./frontend/src/assets/output.css --watch