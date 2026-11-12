include .env
export

DB_URL = $(DATABASE_URL)

migrate-status:
	goose -dir migrations postgres "$(DB_URL)" status

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

migrate-create:
	goose -dir migrations -s create $(name) sql

.PHONY: fmt lint vet test check

fmt:
	golangci-lint fmt

lint:
	golangci-lint run

vet:
	go vet ./...

test:
	go test -race ./...

# Run everything before a commit
check: vet lint test
