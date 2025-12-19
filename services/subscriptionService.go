package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"go-lang-jwt/database"
	"go-lang-jwt/models"

	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SubscriptionWithSource combines subscription and source data for API response
type SubscriptionWithSource struct {
	Subscription models.Subscription `json:"subscription"`
	Source       models.Source       `json:"source"`
}

// AddSubscription adds a new subscription for a user
func AddSubscription(ctx context.Context, userID string, urlString string) (*models.Subscription, error) {
	// Step 1: Validate URL format
	parsedURL, err := url.ParseRequestURI(urlString)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, errors.New("invalid URL format")
	}

	// Normalize URL (remove trailing slash)
	normalizedURL := parsedURL.String()
	if normalizedURL[len(normalizedURL)-1] == '/' {
		normalizedURL = normalizedURL[:len(normalizedURL)-1]
	}

	// Get collections
	sourceCollection := database.OpenCollection(database.Client, "sources")
	subscriptionCollection := database.OpenCollection(database.Client, "subscriptions")

	// Step 2: Check if source already exists
	var source models.Source
	err = sourceCollection.FindOne(ctx, bson.M{"url": normalizedURL}).Decode(&source)

	if err == mongo.ErrNoDocuments {
		// Source doesn't exist, create new one
		source = models.Source{
			ID:               primitive.NewObjectID(),
			URL:              normalizedURL,
			Name:             parsedURL.Host, // Use hostname as default name
			Status:           models.SourceStatusActive,
			LastCrawledAt:    nil,
			LastAttemptAt:    nil,
			LastError:        "",
			RSSUrl:           "",
			SitemapUrl:       "",
			ETag:             "",
			LastModified:     "",
			TotalArticles:    0,
			SuccessfulCrawls: 0,
			FailedCrawls:     0,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		_, err = sourceCollection.InsertOne(ctx, source)
		if err != nil {
			return nil, fmt.Errorf("failed to create source: %v", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to query source: %v", err)
	}
	// If no error, source already exists and was decoded

	// Step 3: Check if user already subscribed to this source
	existingSubscription := subscriptionCollection.FindOne(ctx, bson.M{
		"user_id":   userID,
		"source_id": source.ID,
	})

	if existingSubscription.Err() == nil {
		return nil, errors.New("already subscribed to this source")
	}

	// Step 4: Create subscription
	subscription := models.Subscription{
		ID:            primitive.NewObjectID(),
		User_id:       userID,
		Source_id:     source.ID,
		Subscribed_at: time.Now(),
	}

	_, err = subscriptionCollection.InsertOne(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %v", err)
	}

	return &subscription, nil
}

// RemoveSubscription removes a user's subscription
func RemoveSubscription(ctx context.Context, userID string, subscriptionID string) error {
	// Step 1: Validate subscription ID format
	objectID, err := primitive.ObjectIDFromHex(subscriptionID)
	if err != nil {
		return errors.New("invalid subscription ID format")
	}

	// Get collection
	subscriptionCollection := database.OpenCollection(database.Client, "subscriptions")

	// Step 2: Find subscription and verify ownership
	var subscription models.Subscription
	err = subscriptionCollection.FindOne(ctx, bson.M{
		"_id":     objectID,
		"user_id": userID,
	}).Decode(&subscription)

	if err == mongo.ErrNoDocuments {
		return errors.New("subscription not found or unauthorized")
	} else if err != nil {
		return fmt.Errorf("failed to query subscription: %v", err)
	}

	// Step 3: Delete subscription
	result, err := subscriptionCollection.DeleteOne(ctx, bson.M{
		"_id":     objectID,
		"user_id": userID,
	})

	if err != nil {
		return fmt.Errorf("failed to delete subscription: %v", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("subscription not found")
	}

	return nil
}

// ListSubscriptions gets all subscriptions for a user with source details
func ListSubscriptions(ctx context.Context, userID string) ([]SubscriptionWithSource, error) {
	// Add this debug log
	log.Printf("DEBUG: Looking for subscriptions with user_id: %s", userID)

	subscriptionCollection := database.OpenCollection(database.Client, "subscriptions")
	sourceCollection := database.OpenCollection(database.Client, "sources")

	// Step 1: Find all subscriptions for user
	cursor, err := subscriptionCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriptions: %v", err)
	}
	defer cursor.Close(ctx)

	// Step 2: Decode subscriptions
	var subscriptions []models.Subscription
	if err = cursor.All(ctx, &subscriptions); err != nil {
		return nil, fmt.Errorf("failed to decode subscriptions: %v", err)
	}

	// Step 3: Fetch source details for each subscription
	var result []SubscriptionWithSource

	for _, subscription := range subscriptions {
		var source models.Source
		err := sourceCollection.FindOne(ctx, bson.M{"_id": subscription.Source_id}).Decode(&source)

		if err != nil {
			// Source might have been deleted, skip this subscription
			continue
		}

		result = append(result, SubscriptionWithSource{
			Subscription: subscription,
			Source:       source,
		})
	}

	return result, nil
}
