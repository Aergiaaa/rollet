migrate-up:
	migrate -database "${DATABASE_URL}" -path internal/database/migrations up
migrate-down:
	migrate -database "${DATABASE_URL}" -path internal/database/migrations down

build:
	@go build -o bin/rollet ./cmd
	