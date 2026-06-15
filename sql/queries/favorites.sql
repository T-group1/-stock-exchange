-- name: AddFavorite :exec
INSERT INTO favorites (user_id, base_currency, quote_currency)
VALUES (@user_id, @base_currency, @quote_currency)
ON CONFLICT (user_id, base_currency, quote_currency) DO NOTHING;

-- name: RemoveFavorite :exec
DELETE FROM favorites
WHERE user_id = @user_id 
  AND base_currency = @base_currency 
  AND quote_currency = @quote_currency;

-- name: GetUserFavorites :many
SELECT base_currency, quote_currency, created_at
FROM favorites
WHERE user_id = @user_id
ORDER BY created_at DESC;

-- name: IsFavorite :one
SELECT EXISTS (
    SELECT 1 FROM favorites 
    WHERE user_id = @user_id 
      AND base_currency = @base_currency 
      AND quote_currency = @quote_currency
);