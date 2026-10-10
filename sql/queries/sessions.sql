-- name: CreateSession :one
INSERT INTO user_sessions (
    user_id,
    session_token_hash,
    expires_at
)
VALUES ($1, $2, $3)
RETURNING *;


-- name: GetSessionByTokenHash :one
SELECT *
FROM user_sessions
WHERE session_token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > NOW()
LIMIT 1;


-- name: RevokeSession :exec
UPDATE user_sessions
SET
    revoked_at = NOW(),
    updated_at = NOW()
WHERE session_token_hash = $1
  AND revoked_at IS NULL;


-- name: RevokeAllUserSessions :exec
UPDATE user_sessions
SET
    revoked_at = NOW(),
    updated_at = NOW()
WHERE user_id = $1
  AND revoked_at IS NULL;


-- name: DeleteExpiredSessions :exec
DELETE FROM user_sessions
WHERE expires_at < NOW();