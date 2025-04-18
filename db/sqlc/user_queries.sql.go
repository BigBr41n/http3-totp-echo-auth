// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: user_queries.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (username, email, password, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id, username, email, password, role, created_at, updated_at
`

type CreateUserParams struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type CreateUserRow struct {
	ID        pgtype.UUID        `json:"id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	Password  string             `json:"password"`
	Role      string             `json:"role"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Username,
		arg.Email,
		arg.Password,
		arg.Role,
	)
	var i CreateUserRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Role,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, username, email, password, role, two_fa_enabled 
FROM users
WHERE email = $1
`

type GetUserByEmailRow struct {
	ID           pgtype.UUID `json:"id"`
	Username     string      `json:"username"`
	Email        string      `json:"email"`
	Password     string      `json:"password"`
	Role         string      `json:"role"`
	TwoFaEnabled pgtype.Bool `json:"two_fa_enabled"`
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i GetUserByEmailRow
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Role,
		&i.TwoFaEnabled,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, username, email, password, role, two_fa_enabled, totp_secret, created_at, updated_at 
FROM users
WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id pgtype.UUID) (User, error) {
	row := q.db.QueryRow(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Role,
		&i.TwoFaEnabled,
		&i.TotpSecret,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const set2FAStatus = `-- name: Set2FAStatus :one
UPDATE users
SET two_fa_enabled = $2
WHERE id = $1
RETURNING id, username, email, password, role, two_fa_enabled, totp_secret, created_at, updated_at
`

type Set2FAStatusParams struct {
	ID           pgtype.UUID `json:"id"`
	TwoFaEnabled pgtype.Bool `json:"two_fa_enabled"`
}

func (q *Queries) Set2FAStatus(ctx context.Context, arg Set2FAStatusParams) (User, error) {
	row := q.db.QueryRow(ctx, set2FAStatus, arg.ID, arg.TwoFaEnabled)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Email,
		&i.Password,
		&i.Role,
		&i.TwoFaEnabled,
		&i.TotpSecret,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const storeSecret2FA = `-- name: StoreSecret2FA :exec
UPDATE users 
SET totp_secret = $2 
WHERE id = $1
`

type StoreSecret2FAParams struct {
	ID         pgtype.UUID `json:"id"`
	TotpSecret pgtype.Text `json:"totp_secret"`
}

func (q *Queries) StoreSecret2FA(ctx context.Context, arg StoreSecret2FAParams) error {
	_, err := q.db.Exec(ctx, storeSecret2FA, arg.ID, arg.TotpSecret)
	return err
}
