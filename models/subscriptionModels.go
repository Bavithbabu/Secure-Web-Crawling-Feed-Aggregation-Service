package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Subscription represents the relationship between a user and a source
type Subscription struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	User_id       string             `bson:"user_id" json:"user_id" validate:"required"`
	Source_id     primitive.ObjectID `bson:"source_id" json:"source_id" validate:"required"`
	Subscribed_at time.Time          `bson:"subscribed_at" json:"subscribed_at"`
}
