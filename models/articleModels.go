package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Article represents a crawled article from a source
type Article struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	Source_id     primitive.ObjectID `bson:"source_id" json:"source_id" validate:"required"`
	Title         string             `bson:"title" json:"title" validate:"required,min=1,max=500"`
	URL           string             `bson:"url" json:"url" validate:"required,url,max=2000"`
	Content_hash  string             `bson:"content_hash" json:"content_hash" validate:"required,len=64"`
	Summary       *string            `bson:"summary" json:"summary" validate:"omitempty,max=1000"`
	Published_at  *time.Time         `bson:"published_at" json:"published_at"`
	Discovered_at time.Time          `bson:"discovered_at" json:"discovered_at"`
	Author        *string            `bson:"author" json:"author" validate:"omitempty,max=200"`
}
