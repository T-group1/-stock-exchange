include .make/openapi.mk
include .make/codegen.mk
include .make/lint.mk

.PHONY: fmt

# Главная команда для глобального форматирования всего проекта
fmt: openapi-format yaml-format