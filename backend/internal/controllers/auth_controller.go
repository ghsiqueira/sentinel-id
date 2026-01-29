package controllers

import (
	"net/http"
	"sentinel-id/internal/models"
	"sentinel-id/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	Service *services.UserService
}

func NewAuthController(service *services.UserService) *AuthController {
	return &AuthController{Service: service}
}

func (c *AuthController) Register(ctx *gin.Context) {
	var input models.UserRegisterDTO

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.Service.Register(input)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user_id": user.ID,
	})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var input models.UserLoginDTO

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userAgent := ctx.GetHeader("User-Agent")
	clientIP := ctx.ClientIP()

	accessToken, refreshToken, err := c.Service.Login(input, userAgent, clientIP)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    900,
	})
}

type RefreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (c *AuthController) Refresh(ctx *gin.Context) {
	var input RefreshInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccessToken, err := c.Service.RefreshToken(input.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"expires_in":   900,
	})
}

func (c *AuthController) Logout(ctx *gin.Context) {
	var input RefreshInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.Service.Logout(input.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

func (c *AuthController) LogoutAll(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	err := c.Service.LogoutAll(userID.(uuid.UUID).String())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke sessions"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "All sessions revoked (Kill Switch activated)"})
}
