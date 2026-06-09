.PHONY: generate generate-api generate-db

# Полная регенерация кода проекта
generate: generate-api generate-db

# Генерация Go-сервера чистым oapi-codegen напрямую из бандла 3.0.3
generate-api: openapi-bundle
	go tool -modfile=./tools/go.mod oapi-codegen -config internal/api/oapi-codegen.yaml $(OPENAPI_BUNDLE)

# Генерация типизированных SQL-запросов
generate-db:
	go tool -modfile=./tools/go.mod sqlc generate -f sqlc.yaml