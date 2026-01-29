package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

var db *pgx.Conn
var rdb *redis.Client

func main() {
	var err error

	dbUrl := "postgres://admin:password123@localhost:5432/sentinel_id"
	db, err = pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(context.Background())
	fmt.Println("PostgreSQL Connected")

	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Redis Connected")

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "online",
			"database": "connected",
			"redis":    "connected",
		})
	})

	r.Run(":8080")
}
