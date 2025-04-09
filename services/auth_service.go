package services

import (
	"context"
	"fmt"
	"time"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/BigBr41n/echoAuth/db/sqlc"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/BigBr41n/echoAuth/utils/jwtImpl"
	"github.com/BigBr41n/echoAuth/utils/validator"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
)

type AuthServiceI interface {
	SignUp(userData *dtos.CreateUserDTO) (pgtype.UUID, error)
	Login(creds *Credentials) (string, string, error)
	RefreshUserToken(reftok string, oldtok string) (string, error)
}

type AuthService struct {
	queries *sqlc.Queries
}

func NewAuthService(qrs *sqlc.Queries) AuthServiceI {
	return &AuthService{
		queries: qrs,
	}
}

type Credentials struct {
	Email    string
	Password string
}

func (usr *AuthService) SignUp(userData *dtos.CreateUserDTO) (pgtype.UUID, error) {

	// validate user input
	err := validator.Validate(userData)
	if err != nil {
		logger.Error("failed to create user",
			zap.String("context", "DTO validation failed"),
			zap.Error(err),
		)
		return pgtype.UUID{}, err
	}
	user, err := usr.queries.CreateUser(context.Background(), (sqlc.CreateUserParams)(*userData))
	if err != nil {
		logger.Error("failed to create user",
			zap.String("reason", err.Error()),
			zap.Error(err),
		)
		return pgtype.UUID{}, err
	}

	logger.Error("new user created",
		zap.String("userId", user.ID.String()),
	)
	return user.ID, nil
}

func (usr *AuthService) Login(creds *Credentials) (string, string, error) {

	user, err := usr.queries.GetUserByEmail(context.Background(), creds.Email)
	if err != nil {
		logger.Error("failed to login",
			zap.String("reason", err.Error()),
			zap.Error(err),
		)
		return "", "", err
	}

	if user.Password != creds.Password {
		logger.Error("failed passwrod checking operation",
			zap.String("user", user.ID.String()),
		)
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
		logger.Error("failed to login",
			zap.String("reason", err.Error()),
			zap.Error(err),
		)
		return "", "", err
	}

	logger.Error("User logged in",
		zap.String("userId", user.ID.String()),
	)
	return accessToken, refreshToken, nil
}

func (usr *AuthService) RefreshUserToken(refTok string, oldTok string) (string, error) {
	newRefTok, err := jwtImpl.RefreshAccessToken(refTok, oldTok)

	if err != nil {
		logger.Error("failed to refresh the token",
			zap.String("reason", err.Error()),
			zap.Error(err),
		)
		return "", err
	}

	return newRefTok, nil
}
