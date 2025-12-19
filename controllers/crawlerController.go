package controllers

import (
	"context"
	"net/http"
	"time"

	"go-lang-jwt/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CrawlSubscription handles POST /api/crawl/:subscription_id
func CrawlSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from JWT
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Get subscription ID from URL
		subscriptionID := c.Param("id")
		if subscriptionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "subscription id required"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get user's subscriptions
		subscriptions, err := services.ListSubscriptions(ctx, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Find the subscription and verify ownership
		var sourceID primitive.ObjectID
		found := false
		for _, sub := range subscriptions {
			if sub.Subscription.ID.Hex() == subscriptionID {
				sourceID = sub.Source.ID
				found = true
				break
			}
		}

		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}

		// Trigger crawl in background goroutine
		go func() {
			crawlCtx, crawlCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer crawlCancel()

			err := services.CrawlSource(crawlCtx, sourceID)
			if err != nil {
				// Log error but don't fail the request
				// User already got "crawl started" response
			}
		}()

		c.JSON(http.StatusOK, gin.H{
			"message": "Crawl started in background",
		})
	}
}

// CrawlAllSubscriptions handles POST /api/crawl/all
func CrawlAllSubscriptions() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user_id from JWT
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Get count of subscriptions
		subscriptions, err := services.ListSubscriptions(ctx, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(subscriptions) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message": "No subscriptions to crawl",
			})
			return
		}

		// Trigger crawl in background goroutine
		go func() {
			crawlCtx, crawlCancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer crawlCancel()

			services.CrawlUserSources(crawlCtx, userID.(string))
		}()

		c.JSON(http.StatusOK, gin.H{
			"message": "Crawl started for all subscriptions in background",
			"count":   len(subscriptions),
		})
	}
}
