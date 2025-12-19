package services

import (
	"context"
	"fmt"

	"go-lang-jwt/database"
	"go-lang-jwt/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FeedArticle combines article with source info
type FeedArticle struct {
	Article models.Article `json:"article"`
	Source  models.Source  `json:"source"`
}

// GetUserFeed returns paginated articles from user's subscribed sources
func GetUserFeed(ctx context.Context, userID string, page int, limit int) ([]FeedArticle, int64, error) {
	// Get user's subscriptions
	subscriptions, err := ListSubscriptions(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subscriptions: %v", err)
	}

	if len(subscriptions) == 0 {
		return []FeedArticle{}, 0, nil
	}

	// Extract source IDs
	var sourceIDs []primitive.ObjectID
	sourceMap := make(map[primitive.ObjectID]models.Source)

	for _, sub := range subscriptions {
		sourceIDs = append(sourceIDs, sub.Source.ID)
		sourceMap[sub.Source.ID] = sub.Source
	}

	// Get articles from subscribed sources
	articleCollection := database.OpenCollection(database.Client, "articles")

	// Count total articles
	totalCount, err := articleCollection.CountDocuments(ctx, bson.M{
		"source_id": bson.M{"$in": sourceIDs},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count articles: %v", err)
	}

	// Calculate pagination
	skip := (page - 1) * limit
	opts := options.Find().
		SetSort(bson.D{{Key: "discovered_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	// Fetch articles
	cursor, err := articleCollection.Find(ctx, bson.M{
		"source_id": bson.M{"$in": sourceIDs},
	}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch articles: %v", err)
	}
	defer cursor.Close(ctx)

	// Decode and combine with source info
	var feed []FeedArticle
	for cursor.Next(ctx) {
		var article models.Article
		if err := cursor.Decode(&article); err != nil {
			continue
		}

		feed = append(feed, FeedArticle{
			Article: article,
			Source:  sourceMap[article.Source_id],
		})
	}

	return feed, totalCount, nil
}
