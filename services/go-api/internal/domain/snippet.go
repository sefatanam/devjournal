package domain

import (
	"time"
)

// Snippet represents a code snippet stored in MongoDB
// Uses flexible schema with metadata for different snippet types
type Snippet struct {
	ID          string                 `json:"id" bson:"_id,omitempty"`
	UserID      string                 `json:"userId" bson:"user_id"`
	Title       string                 `json:"title" bson:"title"`
	Description string                 `json:"description" bson:"description"`
	Code        string                 `json:"code" bson:"code"`
	Language    string                 `json:"language" bson:"language"` // typescript, go, python, etc.
	Tags        []string               `json:"tags" bson:"tags"`
	Metadata    map[string]interface{} `json:"metadata" bson:"metadata"` // Flexible fields
	IsPublic    bool                   `json:"isPublic" bson:"is_public"`
	ViewsCount  int                    `json:"viewsCount" bson:"views_count"`
	CreatedAt   time.Time              `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time              `json:"updatedAt" bson:"updated_at"`
}

// NewSnippet creates a new snippet with timestamps
func NewSnippet(userID, title, description, code, language string, tags []string, metadata map[string]interface{}, isPublic bool) *Snippet {
	now := time.Now().UTC()
	if tags == nil {
		tags = []string{}
	}
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	return &Snippet{
		UserID:      userID,
		Title:       title,
		Description: description,
		Code:        code,
		Language:    language,
		Tags:        tags,
		Metadata:    metadata,
		IsPublic:    isPublic,
		ViewsCount:  0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreateSnippetRequest represents the request to create a snippet
type CreateSnippetRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Code        string                 `json:"code"`
	Language    string                 `json:"language"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsPublic    bool                   `json:"isPublic"`
}

// UpdateSnippetRequest represents the request to update a snippet
type UpdateSnippetRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Code        string                 `json:"code"`
	Language    string                 `json:"language"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	IsPublic    bool                   `json:"isPublic"`
}
