goose-status:
	goose -dir internal/storage/migrations postgres "$(DATABASE_URL)" status

goose-up:
	goose -dir internal/storage/migrations postgres "$(DATABASE_URL)" up

goose-down:
	goose -dir internal/storage/migrations postgres "$(DATABASE_URL)" down

database-up:
	docker compose -f docker-compose.dev.yml up

server:
	go run cmd/main.go
