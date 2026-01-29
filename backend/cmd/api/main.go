package main

import (
	"log"
	"net/http"
	"os"
	_ "sentinel-id/docs"
	"sentinel-id/internal/controllers"
	"sentinel-id/internal/database"
	"sentinel-id/internal/middlewares"
	"sentinel-id/internal/repositories"
	"sentinel-id/internal/services"

	"github.com/gin-contrib/cors"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Sentinel ID API
// @version         1.0
// @description     Sistema de Identidade Centralizado (SSO).
// @host            localhost:8080
// @BasePath        /api
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

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
