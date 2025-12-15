package main

import (
	"fmt"
	"log"
	"os"

	"go-lang-jwt/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Starting JWT Authentication Server...")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	// Public routes (no authentication required)
	routes.AuthRoutes(router)

	// Protected routes (authentication required)
	routes.UserRoutes(router)

	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-1"})
	})

	router.GET("/api-2", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	fmt.Printf("Server is running on port %s\n", port)

	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
