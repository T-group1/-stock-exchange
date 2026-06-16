.PHONY: yaml-format yaml-lint

# Форматирование всех остальных YAML-файлов проекта через инструмент Go
yaml-format:
	go tool -modfile=./tools/go.mod yamlfmt

# Проверка форматирования без изменения файлов
yaml-lint:
	go tool -modfile=./tools/go.mod yamlfmt -lint