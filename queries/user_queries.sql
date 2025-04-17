-- name: CreateUser :one
INSERT INTO users (username, email, password, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, username, email, password, role, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, username, email, password, role, two_fa_enabled 
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * 
FROM users
WHERE id = $1;


-- name: Set2FAStatus :one
UPDATE users
SET two_fa_enabled = $2
WHERE id = $1
RETURNING *;


-- name: StoreSecret2FA :exec 
UPDATE users 
SET totp_secret = $2 
WHERE id = $1; 


