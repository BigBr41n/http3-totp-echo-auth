package services

import (
	"fmt"
	"time"

	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/BigBr41n/echoAuth/models"
	"github.com/BigBr41n/echoAuth/utils/jwtImpl"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserServiceI interface {
	SignUp(userData *models.User) (uuid.UUID, error)
	Login(creds Credentials) (string, string, error)
}

type UserService struct {
	userRepo models.UserRepoI
}

func NewUserService(userRepo models.UserRepoI) UserServiceI {
	return &UserService{
		userRepo,
	}
}

type Credentials struct {
	Email    string
	Password string
}

func (usr *UserService) SignUp(creds *models.User) (uuid.UUID, error) {
	err := usr.userRepo.CreateUser(creds)
	if err != nil {
		logger.Error(err)
		return uuid.Nil, fmt.Errorf("failed to create user")
	}

	return creds.ID, nil
}

func (usr *UserService) Login(creds Credentials) (string, string, error) {

	user, err := usr.userRepo.GetUserByEmail(creds.Email)
	if err != nil {
		return "", "", err
	}

	if user.Password != creds.Password {
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
