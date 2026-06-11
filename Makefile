.PHONY: fmt build test run lint clean generate

# Сборка Go-приложения
build:
	go build -o bin/server ./cmd/server

# Запуск сервера
run: build
	./bin/server

# Тесты
test:
	go test -v -race -coverprofile=coverage.out ./...

# Генерация кода
generate: generate-api generate-db

# Полная проверка
check: yaml-lint test

# Очистка
clean:
	rm -rf bin/ coverage.out