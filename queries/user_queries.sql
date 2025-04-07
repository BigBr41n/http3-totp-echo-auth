-- name: CreateUser :one
INSERT INTO users (username, email, password, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, username, email, password, role, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, username, email, password, role, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, username, email, password, role, created_at, updated_at
FROM users
WHERE id = $1;

