include .env
export

run_code:
	@go run cmd/app/main.go

start_docker:
	@docker compose up --build -d

start_docker_debug:
	@docker compose up --build

end_docker:
	@docker compose down