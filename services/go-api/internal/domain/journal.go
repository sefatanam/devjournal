package domain

import (
	"time"

	"github.com/google/uuid"
)

// JournalEntry represents a learning journal entry
type JournalEntry struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Mood      string    `json:"mood"` // excited, productive, frustrated, confused, accomplished
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// NewJournalEntry creates a new journal entry with generated ID and timestamps
func NewJournalEntry(userID uuid.UUID, title, content, mood string, tags []string) *JournalEntry {
	now := time.Now().UTC()
	if tags == nil {
		tags = []string{}
	}
	return &JournalEntry{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Content:   content,
		Mood:      mood,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// CreateJournalEntryRequest represents the request to create a journal entry
type CreateJournalEntryRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Mood    string   `json:"mood"`
	Tags    []string `json:"tags"`
}

// UpdateJournalEntryRequest represents the request to update a journal entry
type UpdateJournalEntryRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Mood    string   `json:"mood"`
	Tags    []string `json:"tags"`
}
