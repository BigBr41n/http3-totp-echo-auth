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
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceI interface {
	SignUp(userData *dtos.CreateUserDTO) (pgtype.UUID, error)
	Login(creds *Credentials) (string, string, error)
	RefreshUserToken(reftok string, oldtok string) (string, error)
	ValidateTOTP(userID pgtype.UUID, TOTP string) (string, string, error)
	Enable2FA(userEmail string, userID pgtype.UUID, enable bool) (string, string, error)
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
		logger.Error("failed password checking operation",
			zap.String("user", user.ID.String()),
		)
		return "", "", &dtos.ApiErr{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_CREDENTIALS",
			Err:     "Invalid email or password",
			Details: nil,
		}
	}

	// if the user has 2fa enabled
	if user.TwoFaEnabled.Bool {
		claims := &jwtImpl.TempTOTPTokenClaims{
			UserID: user.ID,
			Role:   user.Role,
			Email:  creds.Email,
			TOTP:   true,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		// generate temp
		tempToken, err := jwtImpl.GenerateTempToken(claims)
		if err != nil {
			logger.Error("failed generate TOTP",
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

		return tempToken, "TOTP", nil
	}

	claims := &jwtImpl.CustomAccessTokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(9 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

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

	logger.Info("User logged in",
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

func (usr *AuthService) Enable2FA(userEmail string, userID pgtype.UUID, enable bool) (string, string, error) {

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
			return "", "", &dtos.ApiErr{
				Status:  404,
				Code:    "USER_NOT_FOUND",
				Err:     "User not found or no update occurred",
				Details: nil,
			}
		}

		return "", "", &dtos.ApiErr{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}

	// genrating the totp
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "goEchoAuthApp",
		AccountName: userEmail,
	})

	if err != nil {
		return "", "", &dtos.ApiErr{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}

	secretKey := key.Secret()
	qrCode := key.URL()

	err = usr.queries.StoreSecret2FA(context.Background(), sqlc.StoreSecret2FAParams{
		ID: userID,
		TotpSecret: pgtype.Text{
			String: secretKey,
			Valid:  true,
		},
	})

	if err != nil {
		return "", "", &dtos.ApiErr{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}

	return secretKey, qrCode, nil
}

func (usr *AuthService) ValidateTOTP(userID pgtype.UUID, TOTP string) (string, string, error) {

	user, err := usr.queries.GetUserByID(context.Background(), userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", &dtos.ApiErr{
				Status:  404,
				Code:    "USER_NOT_FOUND",
				Err:     "User not found or no update occurred",
				Details: nil,
			}
		}
		return "", "", &dtos.ApiErr{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}
	// validate the totp
	valid := totp.Validate(TOTP, user.TotpSecret.String)
	if !valid {
		logger.Info("Invalid login", zap.String("userID", userID.String()), zap.String("user OTP", TOTP))
		return "", "", &dtos.ApiErr{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_TOTP",
			Err:     "Invalid TOTP code",
			Details: nil,
		}
	}

	claims := &jwtImpl.CustomAccessTokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(9 * time.Hour)),
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
