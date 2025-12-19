package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SourceStatus string

const (
	SourceStatusActive      SourceStatus = "active"
	SourceStatusError       SourceStatus = "error"
	SourceStatusUnreachable SourceStatus = "unreachable"
)

type Source struct {
	ID     primitive.ObjectID `bson:"_id" json:"id"`
	URL    string             `bson:"url" json:"url" validate:"required,url"`
	Name   string             `bson:"name" json:"name" validate:"required,min=1,max=200"`
	Status SourceStatus       `bson:"status" json:"status"`

	// Crawling metadata
	LastCrawledAt *time.Time `bson:"last_crawled_at" json:"last_crawled_at"`
	LastAttemptAt *time.Time `bson:"last_attempt_at" json:"last_attempt_at"`
	LastError     string     `bson:"last_error" json:"last_error"`

	// Change detection helpers
	RSSUrl       string `bson:"rss_url" json:"rss_url"`
	SitemapUrl   string `bson:"sitemap_url" json:"sitemap_url"`
	ETag         string `bson:"etag" json:"etag"`
	LastModified string `bson:"last_modified" json:"last_modified"`

	// Statistics
	TotalArticles    int `bson:"total_articles" json:"total_articles"`
	SuccessfulCrawls int `bson:"successful_crawls" json:"successful_crawls"`
	FailedCrawls     int `bson:"failed_crawls" json:"failed_crawls"`

	// Timestamps
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
