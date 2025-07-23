run:
	@echo "Запуск сервиса аутентификации:"
	@docker compose up -d

stop:
	@echo "Остановка сервиса аутентификации:"
	@docker compose down

fmt:
	@go fmt ./...

vet: fmt
	@go vet ./...

# при выполнении тестов должен быть запущен Docker

test-handlers: vet
	@echo "Запуск тестов для handlers:"
	@go test -v ./internal/handlers/...

test-services: vet
	@echo "Запуск тестов для services:"
	@go test -v ./internal/services/...

test-database: vet
	@echo "Запуск тестов для database:"
	@go test -v ./internal/storage/database/...

test-memory: vet
	@echo "Запуск тестов для memory:"
	@go test -v ./internal/storage/memory/...

test-cover:
	@go test -cover ./...

test-clean:
	@go clean -testcache