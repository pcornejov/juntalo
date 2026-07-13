.PHONY: run dev test lint check migrate-up migrate-down sqlc-generate seed

run:
	go run ./cmd/api

dev:
	air

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

check: lint test

migrate-up:
	migrate -database "$$DATABASE_URL" -path db/migrations up

migrate-down:
	migrate -database "$$DATABASE_URL" -path db/migrations down 1

sqlc-generate:
	sqlc generate

seed:
	psql "$$DATABASE_URL" -f db/seeds/dev.sql
