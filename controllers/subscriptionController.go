package controllers

import (
	"context"
	"net/http"
	"time"

	"go-lang-jwt/services"

	"github.com/gin-gonic/gin"
)

// AddSubscription handles POST /api/subscriptions
func AddSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from JWT token (set by authentication middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Parse request body
		var req struct {
			URL string `json:"url" binding:"required"`
		}

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "url field is required"})
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Call service
		subscription, err := services.AddSubscription(ctx, userID.(string), req.URL)
		if err != nil {
			if err.Error() == "invalid URL format" {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err.Error() == "already subscribed to this source" {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":      "Subscription created successfully",
			"subscription": subscription,
		})
	}
}

// RemoveSubscription handles DELETE /api/subscriptions/:id
func RemoveSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from JWT token
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Get subscription ID from URL parameter
		subscriptionID := c.Param("id")
		if subscriptionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "subscription id is required"})
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Call service
		err := services.RemoveSubscription(ctx, userID.(string), subscriptionID)
		if err != nil {
			if err.Error() == "invalid subscription ID format" {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err.Error() == "subscription not found or unauthorized" {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Subscription removed successfully",
		})
	}
}

// GetSubscriptions handles GET /api/subscriptions
func GetSubscriptions() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from JWT token
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Call service
		subscriptions, err := services.ListSubscriptions(ctx, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":         len(subscriptions),
			"subscriptions": subscriptions,
		})
	}
}
