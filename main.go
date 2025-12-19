package main

import (
	"fmt"
	"log"
	"os"

	"go-lang-jwt/database"
	"go-lang-jwt/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Starting JWT Authentication Server...")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	// Create indexes after database connection
	if err := database.EnsureIndexes(); err != nil {
		log.Fatal("Failed to create indexes: ", err)
	}

	router := gin.New()
	router.Use(gin.Logger())

	// Public routes (no authentication required)
	routes.AuthRoutes(router)

	// Protected routes (authentication required)
	routes.UserRoutes(router)
	routes.SubscriptionRoutes(router)

	// ADD THIS DEBUG CODE:
	fmt.Println("\n=== Registered Routes ===")
	for _, route := range router.Routes() {
		log.Printf("%s %s", route.Method, route.Path)
	}
	fmt.Println("========================\n")

	fmt.Printf("Server is running on port %s\n", port)

	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
