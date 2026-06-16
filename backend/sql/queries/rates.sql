-- name: CreateRate :one
INSERT INTO currency_rates (currency_code, rate, rate_date, source, change_percentage)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (currency_code, rate_date) DO UPDATE SET
  rate = EXCLUDED.rate,
  source = EXCLUDED.source,
  change_percentage = EXCLUDED.change_percentage
RETURNING *;

-- name: GetLatestRates :many
SELECT DISTINCT ON (currency_code) *
FROM currency_rates
ORDER BY currency_code, rate_date DESC;

-- name: GetRateHistory :many
SELECT rate, rate_date
FROM currency_rates
WHERE currency_code = $1 AND rate_date >= $2
ORDER BY rate_date ASC;