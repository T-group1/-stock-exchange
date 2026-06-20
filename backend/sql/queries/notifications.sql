-- name: CreateNotification :one
INSERT INTO notifications (user_id, subscription_id, type, title, message)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserNotificationsUnread :many
SELECT * FROM notifications
WHERE user_id = $1 AND is_read = false
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserNotificationsCount :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1;

-- name: MarkNotificationAsRead :one
UPDATE notifications
SET is_read = true
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetUnreadCount :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1 AND is_read = false;