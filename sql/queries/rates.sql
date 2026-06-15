-- name: GetLatestRate :one
SELECT currency_code, rate, rate_date
FROM currency_rates
WHERE currency_code = @currency_code
ORDER BY rate_date DESC
LIMIT 1;

-- name: GetRateHistory :many
SELECT currency_code, rate, rate_date
FROM currency_rates
WHERE currency_code = @currency_code
  AND rate_date >= @start_date
  AND rate_date <= @end_date
ORDER BY rate_date ASC;

-- name: GetLatestCrossRate :one
-- Магия: база сама делит курс базовой валюты на курс целевой за последнюю доступную дату
SELECT 
    (r_base.rate / r_quote.rate) AS cross_rate,
    r_base.rate_date
FROM currency_rates r_base
JOIN currency_rates r_quote 
    ON r_base.rate_date = r_quote.rate_date 
    AND r_quote.currency_code = @quote_currency
WHERE r_base.currency_code = @base_currency
ORDER BY r_base.rate_date DESC
LIMIT 1;

-- name: GetCrossRateHistory :many
-- Отдает готовую историю кросс-курса по дням. Бэкендер просто строит по этому график.
SELECT 
    r_base.rate_date,
    (r_base.rate / r_quote.rate) AS cross_rate
FROM currency_rates r_base
JOIN currency_rates r_quote 
    ON r_base.rate_date = r_quote.rate_date 
    AND r_quote.currency_code = @quote_currency
WHERE r_base.currency_code = @base_currency
  AND r_base.rate_date >= @start_date
  AND r_base.rate_date <= @end_date
ORDER BY r_base.rate_date ASC;