package main

import (
	"log"
	"sentinel-id/internal/controllers"
	"sentinel-id/internal/database"
	"sentinel-id/internal/repositories"
	"sentinel-id/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()
	defer database.CloseDB()

	userRepo := repositories.NewUserRepository(database.DB)
	userService := services.NewUserService(userRepo)
	authController := controllers.NewAuthController(userService)

	r := gin.Default()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
		}
	}

	log.Println("Server running on port 8080")
	r.Run(":8080")
}
