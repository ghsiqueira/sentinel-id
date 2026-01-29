package main

import (
	"log"
	"net/http"
	"sentinel-id/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()
	defer database.CloseDB()

	r := gin.Default()

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "active",
			"system": "Sentinel ID",
		})
	})

	log.Println("Server running on port 8080")
	r.Run(":8080")
}
