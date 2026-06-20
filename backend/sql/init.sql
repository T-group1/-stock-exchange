-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_verified BOOLEAN DEFAULT false,
    verification_token VARCHAR(255),
    verification_token_expires BIGINT,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT
);

-- Таблица валют
CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    nominal INT NOT NULL DEFAULT 1
);

-- Таблица курсов валют (адаптирована под существующую структуру)
CREATE TABLE IF NOT EXISTS currency_rates (
    id BIGSERIAL PRIMARY KEY,
    currency_code VARCHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE CASCADE,
    rate NUMERIC(10, 4) NOT NULL,
    rate_date DATE NOT NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'cb_rf',
    change_percentage NUMERIC(10, 4) DEFAULT 0,
    UNIQUE(currency_code, rate_date)
);

-- Таблица подписок (алертов)
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL,
    rate_value NUMERIC(10, 4) NOT NULL,
    condition VARCHAR(10) NOT NULL CHECK (condition IN ('above', 'below')),
    is_active BOOLEAN DEFAULT true,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    triggered_at BIGINT
);

-- Таблица уведомлений
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('rate_alert', 'system', 'info')),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT false,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT
);

-- Таблица настроек уведомлений
CREATE TABLE IF NOT EXISTS notification_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email_enabled BOOLEAN DEFAULT true,
    browser_enabled BOOLEAN DEFAULT true,
    quiet_hours_start VARCHAR(5),
    quiet_hours_end VARCHAR(5)
);

-- Таблица избранных валютных пар
CREATE TABLE IF NOT EXISTS favorites (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency_pair VARCHAR(7) NOT NULL, -- Формат: "USD_RUB" (7 символов)
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    PRIMARY KEY (user_id, currency_pair)
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_active ON subscriptions(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_subscriptions_currency ON subscriptions(currency_code);
CREATE INDEX IF NOT EXISTS idx_currency_rates_history ON currency_rates(currency_code, rate_date DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read);