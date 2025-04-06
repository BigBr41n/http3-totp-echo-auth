package jwtImpl

import (
	"errors"
	"time"

	"github.com/BigBr41n/echoAuth/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CustomAccessTokenClaims struct {
	UserID pgtype.UUID `json:"user_id"`
	Role   pgtype.Text `json:"role"`
	Email  string      `json:"email"`
	jwt.RegisteredClaims
}

type CustomRefreshTokenClaims struct {
	UserID pgtype.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

var (
	jwt_sec     string = config.AppConfig.JWTSEC
	jwt_ref_sec string = config.AppConfig.JWTREFSEC
)

func GenerateToken(data *CustomAccessTokenClaims) (string, string, error) {

	/*claims := jwt.MapClaims{
		"sub": user_id,
		"exp": time.Now().Add(time.Hour * 2).Unix(),
	} */
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	signedToken, err := accessToken.SignedString([]byte(jwt_sec))
	if err != nil {
		return "", "", err
	}

	// refresh token sign
	refreshTokenClaims := CustomRefreshTokenClaims{
		UserID: data.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	signedRefToken, err := refreshToken.SignedString([]byte(jwt_ref_sec))
	if err != nil {
		return "", "", err
	}

	return signedToken, signedRefToken, nil
}

func RefreshAccessToken(reftok string, data *CustomAccessTokenClaims) (string, error) {
	parsedToken, err := jwt.ParseWithClaims(reftok, &CustomRefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwt_ref_sec), nil
	})

	if err != nil || !parsedToken.Valid {
		return "", errors.New("invalid or expired refresh token")
	}

	claims, ok := parsedToken.Claims.(*CustomRefreshTokenClaims)
	if !ok {
		return "", errors.New("invalid refresh token claims")
	}

	// Generate new access token
	newAccessTokenClaims := CustomAccessTokenClaims{
		UserID: claims.UserID,
		Role:   data.Role,
		Email:  data.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)), // 15 minutes expiration
		},
	}

	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessTokenClaims)
	return newAccessToken.SignedString([]byte(jwt_sec))
}
