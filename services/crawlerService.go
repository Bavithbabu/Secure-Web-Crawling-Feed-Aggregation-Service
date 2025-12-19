package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"go-lang-jwt/database"
	"go-lang-jwt/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CrawlSource crawls a single source and saves articles
func CrawlSource(ctx context.Context, sourceID primitive.ObjectID) error {
	sourceCollection := database.OpenCollection(database.Client, "sources")
	articleCollection := database.OpenCollection(database.Client, "articles")

	// Step 1: Get source details
	var source models.Source
	err := sourceCollection.FindOne(ctx, bson.M{"_id": sourceID}).Decode(&source)
	if err != nil {
		return fmt.Errorf("source not found: %v", err)
	}

	// Step 2: Update last attempt timestamp
	now := time.Now()
	sourceCollection.UpdateOne(ctx, bson.M{"_id": sourceID}, bson.M{
		"$set": bson.M{"last_attempt_at": now},
	})

	// Step 3: Extract articles from URL
	log.Printf("Crawling source: %s (%s)", source.Name, source.URL)
	articles, err := ExtractArticles(ctx, source)
	if err != nil {
		// Update source with error
		sourceCollection.UpdateOne(ctx, bson.M{"_id": sourceID}, bson.M{
			"$set": bson.M{
				"status":     models.SourceStatusError,
				"last_error": err.Error(),
				"updated_at": time.Now(),
			},
			"$inc": bson.M{"failed_crawls": 1},
		})
		return fmt.Errorf("failed to extract articles: %v", err)
	}

	log.Printf("Found %d articles from %s", len(articles), source.Name)

	// Step 4: Save articles (deduplicate)
	savedCount := 0
	for _, articleData := range articles {
		// Check if article already exists (by source_id + url)
		existingCount, _ := articleCollection.CountDocuments(ctx, bson.M{
			"source_id": sourceID,
			"url":       articleData.URL,
		})

		if existingCount > 0 {
			log.Printf("Skipping duplicate URL: %s", articleData.URL)
			continue
		}

		// Check if content hash already exists (duplicate content)
		hashCount, _ := articleCollection.CountDocuments(ctx, bson.M{
			"content_hash": articleData.ContentHash,
		})

		if hashCount > 0 {
			log.Printf("Skipping duplicate content: %s", articleData.Title)
			continue
		}

		// Create article document
		var summary *string
		if articleData.Summary != "" {
			summary = &articleData.Summary
		}

		var author *string
		if articleData.Author != "" {
			author = &articleData.Author
		}

		article := models.Article{
			ID:            primitive.NewObjectID(),
			Source_id:     sourceID,
			Title:         articleData.Title,
			URL:           articleData.URL,
			Content_hash:  articleData.ContentHash,
			Summary:       summary,
			Published_at:  articleData.PublishedAt,
			Discovered_at: time.Now(),
			Author:        author,
		}

		_, err := articleCollection.InsertOne(ctx, article)
		if err != nil {
			log.Printf("Failed to save article: %v", err)
			continue
		}

		savedCount++
	}

	log.Printf("Saved %d new articles from %s", savedCount, source.Name)

	// Step 5: Cleanup old articles (keep only 50 newest per source)
	err = cleanupOldArticles(ctx, sourceID)
	if err != nil {
		log.Printf("Warning: Failed to cleanup old articles: %v", err)
	}

	// Step 6: Update source with success
	sourceCollection.UpdateOne(ctx, bson.M{"_id": sourceID}, bson.M{
		"$set": bson.M{
			"status":          models.SourceStatusActive,
			"last_crawled_at": now,
			"last_error":      "",
			"updated_at":      time.Now(),
		},
		"$inc": bson.M{
			"successful_crawls": 1,
			"total_articles":    savedCount,
		},
	})

	return nil
}

// cleanupOldArticles keeps only the 50 newest articles per source
func cleanupOldArticles(ctx context.Context, sourceID primitive.ObjectID) error {
	articleCollection := database.OpenCollection(database.Client, "articles")

	// Count total articles for this source
	count, err := articleCollection.CountDocuments(ctx, bson.M{"source_id": sourceID})
	if err != nil {
		return err
	}

	// If less than 50, no cleanup needed
	if count <= 50 {
		return nil
	}

	// Find articles to delete (oldest ones)
	opts := options.Find().
		SetSort(bson.D{{Key: "discovered_at", Value: -1}}).
		SetSkip(50)

	cursor, err := articleCollection.Find(ctx, bson.M{"source_id": sourceID}, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var articlesToDelete []primitive.ObjectID
	for cursor.Next(ctx) {
		var article models.Article
		if err := cursor.Decode(&article); err != nil {
			continue
		}
		articlesToDelete = append(articlesToDelete, article.ID)
	}

	// Delete old articles
	if len(articlesToDelete) > 0 {
		_, err = articleCollection.DeleteMany(ctx, bson.M{
			"_id": bson.M{"$in": articlesToDelete},
		})
		if err != nil {
			return err
		}
		log.Printf("Deleted %d old articles from source", len(articlesToDelete))
	}

	return nil
}

// CrawlUserSources crawls all sources for a specific user
func CrawlUserSources(ctx context.Context, userID string) (int, int, error) {
	// Get user's subscriptions
	subscriptions, err := ListSubscriptions(ctx, userID)
	if err != nil {
		return 0, 0, err
	}

	successCount := 0
	failCount := 0

	// Crawl each subscribed source
	for _, sub := range subscriptions {
		err := CrawlSource(ctx, sub.Source.ID)
		if err != nil {
			log.Printf("Failed to crawl source %s: %v", sub.Source.Name, err)
			failCount++
		} else {
			successCount++
		}
	}

	return successCount, failCount, nil
}
