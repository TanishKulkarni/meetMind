package main

import (
	"log"
	"os"

	"meet-ai/backend/config"
	"meet-ai/backend/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Println(".env file not found, using system environment variables")
	}

	db := config.ConnectDatabase()

	defer db.Close()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(router, db)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)

	router.Run(":" + port)
}
