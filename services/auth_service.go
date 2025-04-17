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
	"github.com/BigBr41n/echoAuth/utils/transaction"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceI interface {
	SignUp(ctx context.Context, userData *dtos.CreateUserDTO) (pgtype.UUID, error)
	Login(ctx context.Context, creds *Credentials) (string, string, error)
	RefreshUserToken(reftok string, oldtok string) (string, error)
	ValidateTOTP(ctx context.Context, userID pgtype.UUID, TOTP string) (string, string, error)
	Enable2FA(ctx context.Context, userEmail string, userID pgtype.UUID, enable bool) (string, string, error)
}

type AuthService struct {
	queries *sqlc.Queries
	db      *pgxpool.Pool
}

func NewAuthService(qrs *sqlc.Queries, pgdb *pgxpool.Pool) AuthServiceI {
	return &AuthService{
		queries: qrs,
		db:      pgdb,
	}
}

type Credentials struct {
	Email    string
	Password string
}

func (usr *AuthService) SignUp(ctx context.Context, userData *dtos.CreateUserDTO) (pgtype.UUID, error) {

	tx, err := transaction.StartTransaction(ctx, usr.db)
	if err != nil {
		logger.Error("error when startsing a transaction",
			zap.String("context", "error in function start transaction from utils"),
			zap.Error(err),
		)

		return pgtype.UUID{}, &dtos.ApiErr{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}
	defer tx.Rollback(ctx)
	qtx := usr.queries.WithTx(tx)

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

	user, err := qtx.CreateUser(ctx, (sqlc.CreateUserParams)(*userData))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation error code
			logger.Warn("email already exists",
				zap.String("email", userData.Email),
				zap.Error(err),
			)
			return pgtype.UUID{}, &dtos.ApiErr{
				Status:  http.StatusConflict,
				Code:    "EMAIL_EXISTS",
				Err:     "email already exists",
				Details: nil,
			}
		}

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

	// commit the transaction
	err = tx.Commit(ctx)
	if err != nil {
		logger.Error("failed to create user",
			zap.String("commit", "failed"),
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

	return user.ID, nil
}

func (usr *AuthService) Login(ctx context.Context, creds *Credentials) (string, string, error) {

	user, err := usr.queries.GetUserByEmail(ctx, creds.Email)
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

func (usr *AuthService) Enable2FA(ctx context.Context, userEmail string, userID pgtype.UUID, enable bool) (string, string, error) {

	// start a transaction
	tx, err := transaction.StartTransaction(ctx, usr.db)
	if err != nil {
		logger.Error("error when startsing a transaction",
			zap.String("context", "error in function start transaction from utils"),
			zap.Error(err),
		)

		return "", "", &dtos.ApiErr{
			Status:  http.StatusInternalServerError,
			Code:    "INTERNAL_ERROR",
			Err:     err.Error(),
			Details: nil,
		}
	}
	defer tx.Rollback(ctx)
	qtx := usr.queries.WithTx(tx)

	// enable 2FA in the DB
	_, err = qtx.Set2FAStatus(ctx, sqlc.Set2FAStatusParams{
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

	err = qtx.StoreSecret2FA(ctx, sqlc.StoreSecret2FAParams{
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

	err = tx.Commit(ctx)
	if err != nil {
		logger.Error("failed to enable 2fa",
			zap.String("commit", "failed"),
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

	return secretKey, qrCode, nil
}

func (usr *AuthService) ValidateTOTP(ctx context.Context, userID pgtype.UUID, TOTP string) (string, string, error) {

	user, err := usr.queries.GetUserByID(ctx, userID)

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
