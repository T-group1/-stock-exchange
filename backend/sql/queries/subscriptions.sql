-- name: CreateSubscription :one
INSERT INTO subscriptions (user_id, currency_code, rate_value, condition)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserSubscriptions :many
SELECT * FROM subscriptions
WHERE user_id = $1;

-- name: GetActiveSubscriptionsForCurrency :many
SELECT * FROM subscriptions
WHERE currency_code = $1 AND is_active = true;

-- name: DeactivateSubscription :exec
UPDATE subscriptions
SET is_active = false, triggered_at = $2
WHERE id = $1;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1;

-- name: GetActiveSubscriptions :many
SELECT * FROM subscriptions
WHERE is_active = true;