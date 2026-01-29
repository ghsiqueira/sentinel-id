package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

var DB *pgxpool.Pool
var RDB *redis.Client

func ConnectDB() {
	var err error

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	var dbUrl string

	if dbHost != "" && dbUser != "" {
		dbUrl = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbPort, dbName)
	} else {
		dbUrl = os.Getenv("DATABASE_URL")
	}

	if dbUrl == "" {
		log.Fatal("Erro: Não foi possível determinar a conexão com o banco (nem variáveis separadas, nem DATABASE_URL)")
	}

	DB, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	err = DB.Ping(context.Background())
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	fmt.Println("PostgreSQL connected successfully")

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	var redisAddr string

	if redisHost != "" {
		redisAddr = fmt.Sprintf("%s:%s", redisHost, redisPort)
	} else {
		redisAddr = os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			redisAddr = "localhost:6379"
		}
	}

	RDB = redis.NewClient(&redis.Options{
		Addr: redisAddr,
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
