package main

import (
	"log"
	"net/http"
	"os"
	"sentinel-id/internal/controllers"
	"sentinel-id/internal/database"
	"sentinel-id/internal/middlewares"
	"sentinel-id/internal/repositories"
	"sentinel-id/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.ConnectDB()
	defer database.CloseDB()

	userRepo := repositories.NewUserRepository(database.DB)
	sessionRepo := repositories.NewSessionRepository(database.DB)

	userService := services.NewUserService(userRepo, sessionRepo)

	authController := controllers.NewAuthController(userService)

	r := gin.Default()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh", authController.Refresh)
			auth.POST("/logout", authController.Logout)
		}

		protected := api.Group("/users")
		protected.Use(middlewares.AuthMiddleware())
		{
			protected.GET("/me", func(c *gin.Context) {
				userID, _ := c.Get("userID")
				c.JSON(http.StatusOK, gin.H{
					"message": "Você entrou na área VIP!",
					"your_id": userID,
				})
			})

			protected.POST("/revoke-all", authController.LogoutAll)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port " + port)
	r.Run(":" + port)
}
