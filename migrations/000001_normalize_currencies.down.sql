-- Откат миграции

-- Удаляем материализованное представление
DROP MATERIALIZED VIEW IF EXISTS current_rates;

-- Создаём старую таблицу
CREATE TABLE rates_old (
    id BIGSERIAL PRIMARY KEY,
    pair VARCHAR(20) NOT NULL,
    rate DECIMAL(20, 6) NOT NULL,
    date DATE NOT NULL,
    UNIQUE(pair, date)
);

-- Миграция обратно в строковый формат
INSERT INTO rates_old (pair, rate, date)
SELECT 
    base_currency || '_' || quote_currency AS pair,
    rate,
    date
FROM rates;

-- Удаляем новую таблицу
DROP TABLE rates;

-- Переименовываем
ALTER TABLE rates_old RENAME TO rates;

-- Удаляем таблицу валют
DROP TABLE IF EXISTS currencies;

-- Удаляем функцию
DROP FUNCTION IF EXISTS SPLIT_PAIR(VARCHAR, INT);