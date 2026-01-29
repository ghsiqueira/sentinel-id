package controllers

import (
	"net/http"
	"sentinel-id/internal/models"
	"sentinel-id/internal/services"

	"github.com/gin-gonic/gin"
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
