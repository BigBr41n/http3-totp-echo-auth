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
	Role   string      `json:"role"`
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

func RefreshAccessToken(reftok string, old string) (string, error) {

	parsedAccToken, _, err := ParseExtractClaims(old, "access", jwt_sec)
	if err != nil {
		return "", err
	}
	accClaims := parsedAccToken.Claims.(*CustomAccessTokenClaims)

	_, valid, err := ParseExtractClaims(old, "refresh", jwt_sec)
	if err != nil || !valid {
		if !valid {
			return "", errors.New("reftok not valid")
		}
		return "", err
	}
	// Generate new access token
	newAccessTokenClaims := CustomAccessTokenClaims{
		UserID: accClaims.UserID,
		Role:   accClaims.Role,
		Email:  accClaims.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)), // 15 minutes expiration
		},
	}

	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newAccessTokenClaims)
	return newAccessToken.SignedString([]byte(jwt_sec))
}

func ParseExtractClaims(tok string, typ string, secret string) (jwt.Token, bool, error) {
	var claims jwt.Claims
	var err error

	if typ == "access" {
		claims = &CustomAccessTokenClaims{}
	} else if typ == "refresh" {
		claims = &CustomRefreshTokenClaims{}
	} else {
		return jwt.Token{}, false, errors.New("invalid token type")
	}

	// Parse the token
	parsedToken, err := jwt.ParseWithClaims(tok, claims, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return jwt.Token{}, false, err // Return any parsing errors
	}

	return *parsedToken, parsedToken.Valid, nil
}
