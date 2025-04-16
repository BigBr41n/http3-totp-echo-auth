package services

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/BigBr41n/echoAuth/db/sqlc"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/BigBr41n/echoAuth/utils/jwtImpl"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

	//var servErr *dtos.ApiErr
	// hashing the password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to create user",
			zap.String("context", "error while hashing the password"),
			zap.Error(err),
		)

		return pgtype.UUID{}, &dtos.ApiErr{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}
	userData.Password = string(hashedPass)

	user, err := usr.queries.CreateUser(context.Background(), (sqlc.CreateUserParams)(*userData))
	if err != nil {
		logger.Error("failed to create user",
			zap.String("reason", err.Error()),
			zap.Error(err),
		)

		return pgtype.UUID{}, &dtos.ApiErr{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
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
		return "", "", &dtos.ApiErr{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERR",
			Err:     err.Error(),
			Details: nil,
		}
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password))

	if err != nil {
		logger.Error("failed passwrod checking operation",
			zap.String("user", user.ID.String()),
		)
		return "", "", &dtos.ApiErr{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_CREDENTIALS",
			Err:     "Invalid email or password",
			Details: nil,
		}
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
		return "", "", &dtos.ApiErr{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
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
		return "", &dtos.ApiErr{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}

	return newRefTok, nil
}

func (usr AuthService) Enable2FA(userID pgtype.UUID, enable bool) error {

	// enable 2FA in the DB
	_, err := usr.queries.Set2FAStatus(context.Background(), sqlc.Set2FAStatusParams{
		ID: userID,
		TwoFaEnabled: pgtype.Bool{
			Bool:  enable,
			Valid: true,
		},
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &dtos.ApiErr{
				Status:  404,
				Code:    "USER_NOT_FOUND",
				Err:     "User not found or no update occurred",
				Details: nil,
			}
		}

		return &dtos.ApiErr{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}

	return nil
}
