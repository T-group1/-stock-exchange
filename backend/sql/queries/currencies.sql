-- name: CreateCurrency :one
INSERT INTO currencies (code, name, nominal)
VALUES ($1, $2, $3)
ON CONFLICT (code) DO UPDATE SET
  name = EXCLUDED.name,
  nominal = EXCLUDED.nominal
RETURNING *;

-- name: GetCurrencies :many
SELECT * FROM currencies
ORDER BY code;

-- name: GetCurrencyByCode :one
SELECT * FROM currencies
WHERE code = $1;