-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT
);

-- Таблица валют
CREATE TABLE IF NOT EXISTS currencies (
    code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    nominal INT NOT NULL DEFAULT 1
);

-- Таблица избранных пар валют (Добавлена!)
CREATE TABLE IF NOT EXISTS favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    base_currency VARCHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE CASCADE,
    quote_currency VARCHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE CASCADE,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    UNIQUE(user_id, base_currency, quote_currency)
);

-- Таблица курсов валют
CREATE TABLE IF NOT EXISTS currency_rates (
    id BIGSERIAL PRIMARY KEY,
    currency_code VARCHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE CASCADE,
    rate NUMERIC(15, 6) NOT NULL, -- Увеличил точность до 6 знаков для корректных кросс-курсов
    rate_date DATE NOT NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'cb_rf',
    change_percentage NUMERIC(10, 4) DEFAULT 0,
    UNIQUE(currency_code, rate_date) -- Этот UNIQUE уже создает индекс, дополнительный не нужен!
);

-- Таблица подписок (алертов)
CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    currency_code VARCHAR(3) NOT NULL REFERENCES currencies(code) ON DELETE CASCADE,
    rate_value NUMERIC(15, 6) NOT NULL,
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

-- === ИНДЕКСЫ ДЛЯ ПРОИЗВОДИТЕЛЬНОСТИ ===
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_active ON subscriptions(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_subscriptions_currency ON subscriptions(currency_code);
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_favorites_user ON favorites(user_id);
-- Индекс для быстрого поиска курсов по дате (если понадобится фильтр только по дате)
CREATE INDEX IF NOT EXISTS idx_currency_rates_date ON currency_rates(rate_date DESC);