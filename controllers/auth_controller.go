package controllers

import (
	"net/http"
	"strings"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/BigBr41n/echoAuth/services"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

type AuthController struct {
	userv services.AuthServiceI
}

type AuthControllerI interface {
	RegisterNewUser(c echo.Context) error
	LoginUser(c echo.Context) error
	RefreshAxsToken(c echo.Context) error
}

func NewAuthController(usrSrv services.AuthServiceI) AuthControllerI {
	return &AuthController{
		userv: usrSrv,
	}
}

func (uc *AuthController) RegisterNewUser(c echo.Context) error {
	var cUserDto dtos.CreateUserDTO
	var err error
	var uuid pgtype.UUID

	// bind the request body
	if err = c.Bind(&cUserDto); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid input")
	}

	// call singup service
	if uuid, err = uc.userv.SignUp(&cUserDto); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "User created successfully",
		"uuid":    uuid,
	})
}

func (uc *AuthController) LoginUser(c echo.Context) error {
	var loUserDTO dtos.LoginUserDTO
	var err error
	var accessTok, refreshTok string

	// bind the body
	if err = c.Bind(&loUserDTO); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid Input")
	}

	// login the user
	if accessTok, refreshTok, err = uc.userv.Login((*services.Credentials)(&loUserDTO)); err != nil {
		return c.JSON(http.StatusUnauthorized, "Invalid email or password")
	}

	// returning tokens
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"message":      "Successfull operation",
		"accessToken":  accessTok,
		"refreshToken": refreshTok,
	})
}

func (uc AuthController) RefreshAxsToken(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	oldToken := c.Request().Header.Get("X-Old-Token")

	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Missing Authorization header",
		})
	}
	if oldToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing old token",
		})
	}

	// Check if it's a Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid Authorization header format",
		})
	}

	token := parts[1]

	newRefTok, err := uc.userv.RefreshUserToken(token, oldToken)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	return c.JSON(http.StatusAccepted, map[string]string{
		"message": newRefTok,
	})
}
