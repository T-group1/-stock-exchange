-- Таблица 1: Справочник доступных валют
CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(3) PRIMARY KEY,      -- ISO-код: USD, EUR, CNY
    name text NOT NULL                -- Полное название
);

-- Таблица 2: История ежедневных курсов
CREATE TABLE IF NOT EXISTS currency_rates (
    id bigserial PRIMARY KEY,
    currency_code VARCHAR(3) REFERENCES currencies(code) ON DELETE CASCADE,
    rate numeric(10, 4) NOT NULL,     
    rate_date DATE NOT NULL,          
    UNIQUE (currency_code, rate_date) 
);

-- Таблица 3: Правила уведомлений
CREATE TABLE IF NOT EXISTS alert_rules (
    id bigserial PRIMARY KEY,
    user_id bigint NOT NULL,          
    currency_code VARCHAR(3) REFERENCES currencies(code) ON DELETE CASCADE,
    target_rate numeric(10, 4) NOT NULL,
    condition_type text NOT NULL,     
    is_active boolean DEFAULT true
);

-- Индекс для графиков
CREATE INDEX IF NOT EXISTS idx_currency_rates_history 
ON currency_rates (currency_code, rate_date DESC);

-- Базовое наполнение
INSERT INTO currencies (code, name) VALUES 
('USD', 'Доллар США'),
('EUR', 'Евро'),
('CNY', 'Китайский юань')
ON CONFLICT (code) DO NOTHING;
