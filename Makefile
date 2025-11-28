## run/web: run the cmd/web application
.PHONY: run/web
run/web:
	@go run ./cmd/web

## build: build the application
.PHONY: build
build:
	@echo "Building application..."
	docker compose build

## up: start all services
.PHONY: up
up:
	@echo "Starting all services..."
	docker compose up -d

## down: stop all services
.PHONY: down
down:
	@echo "Stopping all services..."
	docker compose down

## logs: show application logs
.PHONY: logs
logs:
	docker compose logs -f go

## test: run tests
.PHONY: test
test:
	@echo "Running tests..."
	docker compose exec go go test ./...

## clean: remove containers and volumes
.PHONY: clean
clean:
	@echo "Cleaning up..."
	docker compose down -v
	@echo "Clean complete!"
