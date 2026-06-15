-- name: UpsertNotificationSettings :one
INSERT INTO notification_settings (user_id, email_enabled, browser_enabled, quiet_hours_start, quiet_hours_end)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id) DO UPDATE SET
  email_enabled = EXCLUDED.email_enabled,
  browser_enabled = EXCLUDED.browser_enabled,
  quiet_hours_start = EXCLUDED.quiet_hours_start,
  quiet_hours_end = EXCLUDED.quiet_hours_end
RETURNING *;

-- name: GetNotificationSettings :one
SELECT * FROM notification_settings
WHERE user_id = $1;