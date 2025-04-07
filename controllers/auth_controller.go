package controllers

import (
	"net/http"

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
		return c.JSON(http.StatusBadRequest, "Internal Server Error")
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
