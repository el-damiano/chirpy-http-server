-- name: CreateToken :one
INSERT INTO refresh_tokens (
	token,
	created_at,
	updated_at,
	user_id,
	expires_at,
	revoked_at
) VALUES (
	$1,
	now(),
	now(),
	$2,
	$3,
	$4
) RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT user_id FROM refresh_tokens
WHERE token = $1
AND expires_at > now()
AND (revoked_at > now() OR revoked_at is null);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET
	updated_at = now(),
	revoked_at = now()
WHERE token = $1;
