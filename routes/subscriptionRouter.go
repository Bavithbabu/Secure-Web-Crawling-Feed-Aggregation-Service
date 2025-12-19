package routes

import (
	"go-lang-jwt/controllers"
	"go-lang-jwt/middleware"

	"github.com/gin-gonic/gin"
)

// SubscriptionRoutes defines all subscription-related routes
func SubscriptionRoutes(incomingRoutes *gin.Engine) {
	// All subscription routes require authentication
	subscriptionGroup := incomingRoutes.Group("/api/subscriptions")
	subscriptionGroup.Use(middleware.Authenticate())
	{
		subscriptionGroup.POST("", controllers.AddSubscription())
		subscriptionGroup.GET("", controllers.GetSubscriptions())
		subscriptionGroup.DELETE("/:id", controllers.RemoveSubscription())
	}

	// Crawl routes
	crawlGroup := incomingRoutes.Group("/api/crawl")
	crawlGroup.Use(middleware.Authenticate())
	{
		crawlGroup.POST("/:id", controllers.CrawlSubscription())
		crawlGroup.POST("/all", controllers.CrawlAllSubscriptions())
	}

	// Feed routes
	feedGroup := incomingRoutes.Group("/api/feed")
	feedGroup.Use(middleware.Authenticate())
	{
		feedGroup.GET("", controllers.GetFeed())
	}
}
