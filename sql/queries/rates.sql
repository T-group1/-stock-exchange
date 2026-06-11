-- name: InsertCurrency :exec
INSERT INTO currencies (code, name)
VALUES ($1, $2)
ON CONFLICT (code) DO NOTHING;

-- name: SaveCurrencyRate :exec
INSERT INTO currency_rates (currency_code, rate, rate_date)
VALUES ($1, $2, $3)
ON CONFLICT (currency_code, rate_date) 
DO UPDATE SET rate = EXCLUDED.rate;

-- name: GetRatesHistory :many
SELECT rate, rate_date 
FROM currency_rates
WHERE currency_code = $1
ORDER BY rate_date ASC;

-- name: GetLatestRates :many
SELECT DISTINCT ON (currency_code) currency_code, rate, rate_date
FROM currency_rates
ORDER BY currency_code, rate_date DESC;

-- name: CreateAlertRule :exec
INSERT INTO alert_rules (user_id, currency_code, target_rate, condition_type)
VALUES ($1, $2, $3, $4);

-- name: GetActiveAlertRules :many
SELECT id, user_id, currency_code, target_rate, condition_type 
FROM alert_rules 
WHERE is_active = true;
