package services

import (
	"context"
	"fmt"
	"time"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/BigBr41n/echoAuth/db/sqlc"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/BigBr41n/echoAuth/utils/jwtImpl"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserServiceI interface {
	SignUp(userData *dtos.CreateUserDTO) (pgtype.UUID, error)
	Login(creds Credentials) (string, string, error)
}

type UserService struct {
	queries *sqlc.Queries
}

func NewUserService(qrs *sqlc.Queries) UserServiceI {
	return &UserService{
		queries: qrs,
	}
}

type Credentials struct {
	Email    string
	Password string
}

func (usr *UserService) SignUp(userData *dtos.CreateUserDTO) (pgtype.UUID, error) {
	user, err := usr.queries.CreateUser(context.Background(), (sqlc.CreateUserParams)(*userData))
	if err != nil {
		logger.Error(err)
		return pgtype.UUID{}, fmt.Errorf("failed to create user")
	}
	return user.ID, nil
}

func (usr *UserService) Login(creds Credentials) (string, string, error) {

	user, err := usr.queries.GetUserByEmail(context.Background(), creds.Email)
	if err != nil {
		logger.Error(err)
		return "", "", err
	}

	if user.Password != creds.Password {
		logger.Error(err)
		return "", "", fmt.Errorf("invalid credentials")
	}

	claims := &jwtImpl.CustomAccessTokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// generate the auth tokens (access & refresh)
	accessToken, refreshToken, err := jwtImpl.GenerateToken(claims)
	if err != nil {
		logger.Error(err)
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
