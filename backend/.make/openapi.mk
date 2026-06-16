OPENAPI_SRC    := api/openapi.yaml
OPENAPI_BUNDLE := api/openapi.bundle.yaml

.PHONY: openapi-format openapi-bundle

# Форматирование исходного дерева файлов спецификации
openapi-format:
	npx openapi-format $(OPENAPI_SRC) -o $(OPENAPI_SRC) --split

# Сборка всех путей и схем в один монолитный файл спецификации 3.0.3
openapi-bundle:
	npx openapi-format $(OPENAPI_SRC) -o $(OPENAPI_BUNDLE)