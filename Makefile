include .env
export

DB_URL = $(DATABASE_URL)

migrate-status:
	goose -dir migrations postgres "$(DB_URL)" status

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

migrate-create:
	goose -dir migrations -s create $(name) sql
