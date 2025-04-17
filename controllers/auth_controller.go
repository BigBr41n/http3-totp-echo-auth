package controllers

import (
	"net/http"
	"os"
	"strings"

	dtos "github.com/BigBr41n/echoAuth/DTOs"
	"github.com/BigBr41n/echoAuth/internal/logger"
	"github.com/BigBr41n/echoAuth/services"
	"github.com/BigBr41n/echoAuth/utils/jwtImpl"
	"github.com/BigBr41n/echoAuth/utils/response"
	"github.com/BigBr41n/echoAuth/utils/validator"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AuthController struct {
	userv services.AuthServiceI
}

type AuthControllerI interface {
	RegisterNewUser(c echo.Context) error
	LoginUser(c echo.Context) error
	RefreshAxsToken(c echo.Context) error
	Enable2FA(c echo.Context) error
	ValidateTOTP(c echo.Context) error
}

type TOTPInput struct {
	TOTP string `json:"totp"`
}

func NewAuthController(usrSrv services.AuthServiceI) AuthControllerI {
	return &AuthController{
		userv: usrSrv,
	}
}

func (uc *AuthController) RegisterNewUser(c echo.Context) error {

	// extract the context
	ctx := c.Request().Context()

	var cUserDto dtos.CreateUserDTO
	var err error
	var uuid pgtype.UUID

	// bind the request body
	if err = c.Bind(&cUserDto); err != nil {
		return response.ErrResp(c, &dtos.ApiErr{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_OR_MISSED_DATA",
			Err:     "Invalid input data",
			Details: nil,
		})
	}

	// validation layer
	err = validator.Validate(&cUserDto)
	if err != nil {
		return response.ErrResp(c, &dtos.ApiErr{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_INPUT_FORMAT",
			Err:     err.Error(),
			Details: nil,
		})
	}

	// call singup service
	if uuid, err = uc.userv.SignUp(ctx, &cUserDto); err != nil {
		return response.ErrResp(c, err)
	}

	return response.ValResp(c, &dtos.ValidResponse{
		Status:  http.StatusCreated,
		Code:    "USER_CREATED",
		Data:    map[string]interface{}{"uuid": uuid},
		Message: "user signed in successfully",
	})
}

func (uc *AuthController) LoginUser(c echo.Context) error {
	// extract the context
	ctx := c.Request().Context()

	var loUserDTO dtos.LoginUserDTO
	var err error
	var accessTok, refreshTok string

	// bind the body
	if err = c.Bind(&loUserDTO); err != nil {
		logger.Error("binding error", zap.Error(err))
		return response.ErrResp(c, &dtos.ApiErr{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_OR_MISSED_DATA",
			Err:     "invalid data",
			Details: nil,
		})
	}

	// login the user
	if accessTok, refreshTok, err = uc.userv.Login(ctx, (*services.Credentials)(&loUserDTO)); err != nil {
		return response.ErrResp(c, err)
	}
	// returning tokens
	return response.ValResp(c, &dtos.ValidResponse{
		Status:  http.StatusAccepted,
		Code:    "LOGGED_IN",
		Message: "user logged in successfully",
		Data: map[string]interface{}{
			"accessToken":  accessTok,
			"refreshToken": refreshTok,
		},
	})
}

func (uc *AuthController) RefreshAxsToken(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	oldToken := c.Request().Header.Get("X-Old-Token")

	authErr := &dtos.ApiErr{
		Status:  http.StatusBadRequest,
		Code:    "MISSED_TOKEN_OR_HEADER",
		Err:     "",
		Details: nil,
	}

	if authHeader == "" {
		authErr.Err = "Autherization header missed"
		return response.ErrResp(c, authErr)
	}
	if oldToken == "" {
		authErr.Err = "Old token missed"
		return response.ErrResp(c, authErr)
	}

	// Check if it's a Bearer token
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		authErr.Err = "Invalid Authorization header format"
		authErr.Code = "INVALID_TOKEN_FORMAT"
		return c.JSON(http.StatusUnauthorized, authErr)
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

func (uc *AuthController) Enable2FA(c echo.Context) error {

	// extract the context
	ctx := c.Request().Context()

	userData := c.Get("User").(*jwtImpl.CustomAccessTokenClaims)
	var secret string
	var qr string
	var err error

	logger.Debug("user enable 2FA test", zap.String("email", userData.Email), zap.String("id", userData.UserID.String()))

	if secret, qr, err = uc.userv.Enable2FA(ctx, userData.Email, userData.UserID, true); err != nil {
		return response.ErrResp(c, err)
	}

	return response.ValResp(c, &dtos.ValidResponse{
		Status:  http.StatusAccepted,
		Code:    "OTP_ENABLED",
		Message: "otp enabled successfully",
		Data: map[string]interface{}{
			"secret": secret,
			"qr":     qr,
		},
	})
}

func (uc *AuthController) ValidateTOTP(c echo.Context) error {

	// extract the context
	ctx := c.Request().Context()

	var TOTP TOTPInput

	tempToken := c.Request().Header.Get("Autherization")
	parsedToken, val, err := jwtImpl.ParseExtractClaims(tempToken, "temp", os.Getenv("JWTTOTP"))
	if err != nil || !val {
		return response.ErrResp(c, &dtos.ApiErr{
			Status:  http.StatusUnauthorized,
			Code:    "INVALID_TOKEN",
			Err:     "Temp token for the session is expired login again",
			Details: nil,
		})
	}

	if err := c.Bind(&TOTP); err != nil {
		return response.ErrResp(c, &dtos.ApiErr{
			Status:  http.StatusBadRequest,
			Code:    "INVALID_OR_MISSED_DATA",
			Err:     "Invalid input data",
			Details: nil,
		})
	}

	claims := parsedToken.Claims.(*jwtImpl.TempTOTPTokenClaims)
	accessTok, refreshTok, err := uc.userv.ValidateTOTP(ctx, claims.UserID, TOTP.TOTP)
	if err != nil {
		return response.ErrResp(c, err)
	}

	return response.ValResp(c, &dtos.ValidResponse{
		Status:  http.StatusAccepted,
		Code:    "VERIFIED",
		Message: "totp verified and logged in successfully",
		Data: map[string]interface{}{
			"accessToken":  accessTok,
			"refreshToken": refreshTok,
		},
	})

}
