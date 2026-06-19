-- name: CreateUser :one
INSERT INTO users (email, name, password_hash, is_verified, verification_token, verification_token_expires)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: VerifyUser :one
UPDATE users
SET is_verified = true,
    verification_token = NULL,
    verification_token_expires = NULL
WHERE verification_token = $1
  AND verification_token_expires > $2
RETURNING *;
-- name: UpdateVerificationToken :one
UPDATE users
SET verification_token = $1,
    verification_token_expires = $2
WHERE id = $3
RETURNING *;