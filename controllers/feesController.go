package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go-lang-jwt/services"

	"github.com/gin-gonic/gin"
)

// GetFeed handles GET /api/feed
func GetFeed() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Get pagination params (default: page 1, limit 20)
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		feed, total, err := services.GetUserFeed(ctx, userID.(string), page, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		totalPages := (int(total) + limit - 1) / limit

		c.JSON(http.StatusOK, gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"articles":    feed,
		})
	}
}
