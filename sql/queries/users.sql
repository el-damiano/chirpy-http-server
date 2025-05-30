-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
	gen_random_uuid(),
	now(),
	now(),
	$1,
	$2
)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users
WHERE 1=1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET email = $1,
	hashed_password = $2
WHERE id = $3
RETURNING *;

-- name: UpgradeUserMembership :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;
