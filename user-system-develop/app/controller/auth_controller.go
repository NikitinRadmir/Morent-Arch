package controller

import (
	"net/http"

	"user-system/app/dto"
	"user-system/app/exceptions"
	"user-system/app/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	authService service.AuthService
}

func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (c *AuthController) Register(ctx *gin.Context) {
	var input dto.RegisterRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.authService.Register(input)
	if err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response)
}

func (c *AuthController) Login(ctx *gin.Context) {
	var input dto.UserLoginRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.authService.Login(input)
	if err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *AuthController) Logout(ctx *gin.Context) {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := c.authService.Logout(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var input dto.RefreshTokenRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.authService.RefreshToken(input.RefreshToken)
	if err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *AuthController) ChangePassword(ctx *gin.Context) {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var input dto.ChangePasswordRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.authService.ChangePassword(userID, input); err != nil {
		c.handleAuthError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

func (c *AuthController) handleAuthError(ctx *gin.Context, err error) {
	switch e := err.(type) {
	case *exceptions.AuthError:
		switch e.Code {
		case exceptions.ErrInvalidCredentials:
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": e.Message,
				"code":  string(e.Code),
			})
		case exceptions.ErrUserAlreadyExists:
			ctx.JSON(http.StatusConflict, gin.H{
				"error": e.Message,
				"code":  string(e.Code),
			})
		case exceptions.ErrInvalidToken, exceptions.ErrTokenExpired:
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error": e.Message,
				"code":  string(e.Code),
			})
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": e.Message,
				"code":  string(e.Code),
			})
		}
	case *exceptions.AccountDisabledError:
		ctx.JSON(http.StatusForbidden, gin.H{
			"error": e.Error(),
			"code":  string(exceptions.ErrAccountDisabled),
		})
	case *exceptions.InvalidTokenError:
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": e.Error(),
			"code":  string(exceptions.ErrInvalidToken),
		})
	case *exceptions.TokenExpiredError:
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": e.Error(),
			"code":  string(exceptions.ErrTokenExpired),
		})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (c *AuthController) getUserIDFromContext(ctx *gin.Context) (uuid.UUID, error) {
	userIDValue, exists := ctx.Get("userID")
	if !exists {
		return uuid.Nil, exceptions.NewAuthError(
			exceptions.ErrInvalidToken,
			"unauthorized",
		)
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, exceptions.NewAuthError(
			exceptions.ErrInvalidToken,
			"invalid user ID",
		)
	}

	return userID, nil
}
