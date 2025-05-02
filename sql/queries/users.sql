-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
	gen_random_uuid(),
	now(),
	now(),
	$1
)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users
WHERE 1=1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;
