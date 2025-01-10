-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, user_id, expires_at)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = $1,
    updated_at = now()
WHERE refresh_tokens.token = $2
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT users.id, users.email, users.created_at, users.updated_at FROM users
INNER JOIN refresh_tokens
ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1
AND refresh_tokens.revoked_at IS NULL
AND refresh_tokens.expires_at >= now();
