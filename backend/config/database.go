package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDatabase() *pgxpool.Pool {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user,
		password,
		host,
		port,
		dbname,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		log.Fatal("Database ping failed:", err)
	}

	log.Println("Database connected successfully")

	return pool
}
