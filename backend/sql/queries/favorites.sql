-- name: AddFavorite :exec
-- Добавление валюты в избранное. 
-- ON CONFLICT DO NOTHING гарантирует, что при повторном вызове не будет ошибки 500.
INSERT INTO favorites (user_id, currency_code)
VALUES ($1, $2)
ON CONFLICT (user_id, currency_code) DO NOTHING;

-- name: RemoveFavorite :exec
-- Удаление валюты из избранного
DELETE FROM favorites
WHERE user_id = $1 AND currency_code = $2;

-- name: GetUserFavorites :many
-- Получение полного списка избранных валют с их названиями (для красивого JSON в API)
SELECT c.code, c.name, c.nominal
FROM favorites f
JOIN currencies c ON f.currency_code = c.code
WHERE f.user_id = $1
ORDER BY f.created_at DESC;

-- name: GetUserFavoriteCodes :many
-- Получение только кодов (удобно, если в API нужно просто проверить, что валюта в избранном)
SELECT currency_code 
FROM favorites 
WHERE user_id = $1;