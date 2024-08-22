goose-status:
	goose -dir internal/storage/migrations postgres "host=localhost port=5433 user=postgres dbname=job_hunter_bot password=pass sslmode=disable" status

goose-up:
	goose -dir internal/storage/migrations postgres "host=localhost port=5433 user=postgres dbname=job_hunter_bot password=pass sslmode=disable" up

goose-down:
	goose -dir internal/storage/migrations postgres "host=localhost port=5433 user=postgres dbname=job_hunter_bot password=pass sslmode=disable" down

database:
	docker compose -f docker-compose.dev.yml up