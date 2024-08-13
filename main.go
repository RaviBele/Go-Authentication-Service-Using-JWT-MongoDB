package main

import (
	database "go-jwt-auth/database"
	routes "go-jwt-auth/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading environment variables")
	}

	database.DBInstance()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func(c *gin.Context) { c.JSON(200, gin.H{"Success": true}) })
	router.GET("/api-2", func(c *gin.Context) { c.JSON(200, gin.H{"Success": true}) })

	router.Run(":" + port)
}
