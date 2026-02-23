package main

import (
	"context"
	"log"
	"os"
	"sentinel-id/internal/controllers"
	"sentinel-id/internal/middlewares"
	"sentinel-id/internal/repositories"
	"sentinel-id/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "sentinel-id/docs"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL is required")
	}

	dbPool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database: ", err)
	}
	defer dbPool.Close()

	userRepo := repositories.NewUserRepository(dbPool)
	sessionRepo := repositories.NewSessionRepository(dbPool)
	auditRepo := repositories.NewAuditRepository(dbPool)
	qrRepo := repositories.NewQRRepository(dbPool)

	userService := services.NewUserService(userRepo, sessionRepo, auditRepo, qrRepo)

	authController := controllers.NewAuthController(userService)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
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

			auth.POST("/qr/init", authController.InitQR)
			auth.GET("/qr/poll/:code", authController.PollQR)

			auth.POST("/prompt/init", authController.InitPrompt)
		}

		users := api.Group("/users")
		users.Use(middlewares.AuthMiddleware(dbPool))
		{
			users.GET("/me", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "You are authenticated", "user_id": c.MustGet("userID")})
			})
			users.POST("/revoke-all", authController.LogoutAll)
			users.GET("/sessions", authController.ListSessions)
			users.DELETE("/sessions/:id", authController.RevokeSession)
			users.GET("/audit-logs", authController.GetAuditLogs)

			users.POST("/qr/approve", authController.ApproveQR)

			users.POST("/trusted-device", authController.SetTrustedDevice)

			users.GET("/prompt/pending", authController.CheckPendingPrompt)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
