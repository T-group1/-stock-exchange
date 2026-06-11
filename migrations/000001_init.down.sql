-- Таблица валют (справочник)
CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица курсов валют (история)
CREATE TABLE IF NOT EXISTS currency_rates (
    id BIGSERIAL PRIMARY KEY,
    currency_code VARCHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE CASCADE,
    rate NUMERIC(10, 4) NOT NULL,
    rate_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (currency_code, rate_date)
);

-- Таблица пользователей (НОВАЯ!)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Таблица правил уведомлений (алертов)
CREATE TABLE IF NOT EXISTS alert_rules (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE CASCADE,
    target_rate NUMERIC(10, 4) NOT NULL,
    condition_type TEXT NOT NULL CHECK (condition_type IN ('above', 'below')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Таблица уведомлений (НОВАЯ!)
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    alert_rule_id BIGINT REFERENCES alert_rules(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_currency_rates_date ON currency_rates(rate_date DESC);
CREATE INDEX IF NOT EXISTS idx_currency_rates_currency ON currency_rates(currency_code, rate_date DESC);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_alert_rules_user_active ON alert_rules(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_alert_rules_currency ON alert_rules(currency_code, is_active);
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications(created_at DESC);

-- Начальные данные (валюты)
INSERT INTO currencies (code, name, symbol) VALUES
    ('USD', 'Доллар США', '$'),
    ('EUR', 'Евро', '€'),
    ('CNY', 'Китайский юань', '¥')
ON CONFLICT (code) DO NOTHING;