-- name: AddFavorite :exec
-- Добавление валютной пары в избранное.
-- ON CONFLICT DO NOTHING гарантирует, что при повторном вызове не будет ошибки 500.
INSERT INTO favorites (user_id, currency_pair)
VALUES ($1, $2)
ON CONFLICT (user_id, currency_pair) DO NOTHING;

-- name: RemoveFavorite :exec
-- Удаление валютной пары из избранного
DELETE FROM favorites
WHERE user_id = $1 AND currency_pair = $2;

-- name: GetUserFavorites :many
-- Получение полного списка избранных валютных пар
SELECT currency_pair
FROM favorites
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetUserFavoritePairs :many
-- Получение только валютных пар (удобно, если в API нужно просто проверить, что пара в избранном)
SELECT currency_pair
FROM favorites
WHERE user_id = $1;