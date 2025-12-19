package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// createSourceIndexes creates indexes for sources collection
func createSourceIndexes(collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Unique index on url (prevent duplicate publisher URLs)
	urlIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "url", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("url_unique"),
	}

	_, err := collection.Indexes().CreateOne(ctx, urlIndex)
	if err != nil {
		return fmt.Errorf("failed to create source url index: %v", err)
	}

	log.Println("✓ Source indexes created successfully")
	return nil
}

// createSubscriptionIndexes creates indexes for subscriptions collection
func createSubscriptionIndexes(collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Compound unique index on (user_id, source_id) - prevent duplicate subscriptions
	userSourceIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "source_id", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("user_source_unique"),
	}

	// Simple index on user_id for fast user subscription lookups
	userIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetName("user_id_idx"),
	}

	// Create both indexes at once
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		userSourceIndex,
		userIndex,
	})
	if err != nil {
		return fmt.Errorf("failed to create subscription indexes: %v", err)
	}

	log.Println("✓ Subscription indexes created successfully")
	return nil
}

// createArticleIndexes creates indexes for articles collection
func createArticleIndexes(collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Compound unique index on (source_id, url) - prevent duplicate URLs per source
	sourceUrlIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "source_id", Value: 1},
			{Key: "url", Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName("source_url_unique"),
	}

	// Simple index on source_id for finding all articles from a source
	sourceIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "source_id", Value: 1}},
		Options: options.Index().SetName("source_id_idx"),
	}

	// Descending index on published_at for sorting newest articles first
	publishedIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "published_at", Value: -1}},
		Options: options.Index().SetName("published_at_desc"),
	}

	// Index on content_hash for duplicate content detection
	contentHashIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "content_hash", Value: 1}},
		Options: options.Index().SetName("content_hash_idx"),
	}

	// Descending index on discovered_at for tracking recent discoveries
	discoveredIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "discovered_at", Value: -1}},
		Options: options.Index().SetName("discovered_at_desc"),
	}

	// Create all indexes at once
	_, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		sourceUrlIndex,
		sourceIndex,
		publishedIndex,
		contentHashIndex,
		discoveredIndex,
	})
	if err != nil {
		return fmt.Errorf("failed to create article indexes: %v", err)
	}

	log.Println("✓ Article indexes created successfully")
	return nil
}

// EnsureIndexes creates all necessary indexes for the application
func EnsureIndexes() error {
	log.Println("Creating database indexes...")

	// Get database collections
	sourceCollection := OpenCollection(Client, "sources")
	subscriptionCollection := OpenCollection(Client, "subscriptions")
	articleCollection := OpenCollection(Client, "articles")

	// Create indexes for each collection
	if err := createSourceIndexes(sourceCollection); err != nil {
		return err
	}

	if err := createSubscriptionIndexes(subscriptionCollection); err != nil {
		return err
	}

	if err := createArticleIndexes(articleCollection); err != nil {
		return err
	}

	log.Println("✓ All indexes created successfully!")
	return nil
}
