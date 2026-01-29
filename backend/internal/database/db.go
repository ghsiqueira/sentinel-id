package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var DB *pgxpool.Pool
var RDB *redis.Client

func ConnectDB() {
	var err error
	dbUrl := "postgres://admin:password123@localhost:5432/sentinel_id"

	DB, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	err = DB.Ping(context.Background())
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	fmt.Println("PostgreSQL connected successfully")

	RDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err = RDB.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Redis connection failed:", err)
	}
	fmt.Println("Redis connected successfully")
}

func CloseDB() {
	DB.Close()
	RDB.Close()
}
