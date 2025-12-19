package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// GenerateContentHash creates SHA-256 hash from title and content
// Used for detecting duplicate articles
func GenerateContentHash(title string, content string) string {
	// Normalize: lowercase + remove extra whitespace
	normalized := strings.ToLower(strings.TrimSpace(title + " " + content))
	normalized = strings.Join(strings.Fields(normalized), " ")

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}
