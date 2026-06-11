# Currency Exchange API

API для просмотра курсов валют, конвертации и построения графиков. Источник данных: ЦБ РФ.

## Быстрый старт

### Требования
- Go 1.22+
- Docker & Docker Compose
- Node.js 20+ (для OpenAPI инструментов)

### Запуск

```bash
# Клонировать репозиторий
git clone https://github.com/T-group1/-stock-exchange.git
cd -stock-exchange

# Запустить в Docker
docker-compose up -d

# Или локально
make generate
go run ./cmd/server