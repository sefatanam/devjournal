package domain

import (
	"time"

	"github.com/google/uuid"
)

// LearningProgress tracks daily learning progress for streak calculation
type LearningProgress struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"userId"`
	Date              time.Time `json:"date"` // Date only (no time component)
	EntriesCount      int       `json:"entriesCount"`
	SnippetsCount     int       `json:"snippetsCount"`
	StreakDays        int       `json:"streakDays"`
	TotalLearningTime int       `json:"totalLearningTime"` // in minutes
	CreatedAt         time.Time `json:"createdAt"`
}

// NewLearningProgress creates a new progress record for a user on a specific date
func NewLearningProgress(userID uuid.UUID, date time.Time) *LearningProgress {
	return &LearningProgress{
		ID:                uuid.New(),
		UserID:            userID,
		Date:              date.Truncate(24 * time.Hour),
		EntriesCount:      0,
		SnippetsCount:     0,
		StreakDays:        0,
		TotalLearningTime: 0,
		CreatedAt:         time.Now().UTC(),
	}
}

// ProgressSummary provides an overview of user's learning progress
type ProgressSummary struct {
	CurrentStreak     int `json:"currentStreak"`
	LongestStreak     int `json:"longestStreak"`
	TotalEntries      int `json:"totalEntries"`
	TotalSnippets     int `json:"totalSnippets"`
	TotalLearningTime int `json:"totalLearningTime"` // in minutes
	ThisWeekEntries   int `json:"thisWeekEntries"`
	ThisMonthEntries  int `json:"thisMonthEntries"`
}
