include .env
export

run_code:
	@go run cmd/app/main.go

start_docker:
	@docker compose up --build -d app database redis

start_docker_debug:
	@docker compose up --build app database redis

stop_docker:
	@docker compose down 

migrate_up:
	@docker compose run --rm migrate

migrate_down:
	@docker compose run --rm migrate -path=/migrations -database=postgres://${DB_USER}:${DB_PASS}@database:5432/${DB_NAME}?sslmode=disable down 1

migrate_version:
	@docker compose run --rm migrate -path=/migrations -database=postgres://${DB_USER}:${DB_PASS}@database:5432/${DB_NAME}?sslmode=disable version