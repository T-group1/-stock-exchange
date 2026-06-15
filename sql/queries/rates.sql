-- name: CreateRate :one
INSERT INTO currency_rates (pair, rate, rate_date, source, change_percentage)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (pair, rate_date) DO UPDATE SET
  rate = EXCLUDED.rate,
  source = EXCLUDED.source,
  change_percentage = EXCLUDED.change_percentage
RETURNING *;

-- name: GetLatestRates :many
SELECT DISTINCT ON (pair) *
FROM currency_rates
ORDER BY pair, rate_date DESC;

-- name: GetRateHistory :many
SELECT rate, rate_date
FROM currency_rates
WHERE pair = $1 AND rate_date >= $2
ORDER BY rate_date ASC;