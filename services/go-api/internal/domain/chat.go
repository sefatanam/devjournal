package domain

import (
	"time"

	"github.com/google/uuid"
)

// StudyGroup represents a chat room for study collaboration
type StudyGroup struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"isPublic"`
	MaxMembers  int       `json:"maxMembers"`
	CreatedBy   uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewStudyGroup creates a new study group
func NewStudyGroup(name, description string, isPublic bool, maxMembers int, createdBy uuid.UUID) *StudyGroup {
	now := time.Now().UTC()
	return &StudyGroup{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		IsPublic:    isPublic,
		MaxMembers:  maxMembers,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// StudyGroupMember represents membership in a study group
type StudyGroupMember struct {
	GroupID     uuid.UUID `json:"groupId"`
	UserID      uuid.UUID `json:"userId"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role"` // owner, admin, member
	JoinedAt    time.Time `json:"joinedAt"`
}

// ChatMessage represents a message in a study group
type ChatMessage struct {
	ID              string    `json:"id"`
	Room            string    `json:"roomId"`
	UserID          string    `json:"userId"`
	UserDisplayName string    `json:"userDisplayName"`
	Content         string    `json:"content"`
	Type            string    `json:"type"` // message, join, leave
	Timestamp       time.Time `json:"timestamp"`
}

// NewChatMessage creates a new chat message
func NewChatMessage(room, userID, displayName, content, msgType string) *ChatMessage {
	return &ChatMessage{
		ID:              uuid.New().String(),
		Room:            room,
		UserID:          userID,
		UserDisplayName: displayName,
		Content:         content,
		Type:            msgType,
		Timestamp:       time.Now().UTC(),
	}
}
