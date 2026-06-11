-- 1. Создаём таблицу валют
CREATE TABLE currencies (
    code CHAR(3) PRIMARY KEY,  -- ISO 4217: USD, EUR, RUB
    name VARCHAR(100) NOT NULL,
    nominal INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Создаём таблицу курсов с внешними ключами
CREATE TABLE rates_new (
    id BIGSERIAL PRIMARY KEY,
    base_currency CHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE RESTRICT,
    quote_currency CHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE RESTRICT,
    rate DECIMAL(20, 6) NOT NULL CHECK (rate > 0),
    date DATE NOT NULL,
    source VARCHAR(20) NOT NULL CHECK (source IN ('cb_rf', 'calculated')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Уникальность: одна пара на одну дату
    CONSTRAINT uq_rates_pair_date UNIQUE (base_currency, quote_currency, date)
);

-- 3. Индексы для производительности
CREATE INDEX idx_rates_base_quote_date ON rates_new(base_currency, quote_currency, date DESC);
CREATE INDEX idx_rates_date ON rates_new(date DESC);
CREATE INDEX idx_rates_source ON rates_new(source);

-- 4. Миграция данных из старой таблицы rates в новую
INSERT INTO currencies (code, name, nominal)
SELECT DISTINCT 
    SPLIT_PART(pair, '_', 1) AS code,
    SPLIT_PART(pair, '_', 1) AS name,
    1 AS nominal
FROM rates
WHERE pair IS NOT NULL
UNION
SELECT DISTINCT 
    SPLIT_PART(pair, '_', 2) AS code,
    SPLIT_PART(pair, '_', 2) AS name,
    1 AS nominal
FROM rates
WHERE pair IS NOT NULL
ON CONFLICT (code) DO NOTHING;

-- Миграция самих курсов
INSERT INTO rates_new (base_currency, quote_currency, rate, date, source)
SELECT 
    SPLIT_PAIR(pair, 1) AS base_currency,
    SPLIT_PAIR(pair, 2) AS quote_currency,
    rate,
    date,
    'cb_rf' AS source
FROM rates;

-- 5. Вспомогательная функция для парсинга пар
CREATE OR REPLACE FUNCTION SPLIT_PAIR(pair_text VARCHAR, part INT)
RETURNS CHAR(3) AS $$
BEGIN
    IF part = 1 THEN
        RETURN SPLIT_PART(pair_text, '_', 1);
    ELSE
        RETURN SPLIT_PART(pair_text, '_', 2);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- 6. Переименовываем таблицы
DROP TABLE rates;
ALTER TABLE rates_new RENAME TO rates;

-- 7. Создаём материализованное представление для текущих курсов
CREATE MATERIALIZED VIEW IF NOT EXISTS current_rates AS
SELECT DISTINCT ON (base_currency, quote_currency)
    base_currency,
    quote_currency,
    rate,
    date,
    source
FROM rates
ORDER BY base_currency, quote_currency, date DESC;

CREATE UNIQUE INDEX idx_current_rates_pair ON current_rates(base_currency, quote_currency);