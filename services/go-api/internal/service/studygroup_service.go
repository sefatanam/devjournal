package service

import (
	"context"
	"fmt"
	"time"

	"devjournal/internal/domain"
	"devjournal/internal/repository/postgres"

	"github.com/google/uuid"
)

// StudyGroupService handles study group business logic
type StudyGroupService struct {
	groupRepo *postgres.StudyGroupRepository
}

// NewStudyGroupService creates a new study group service
func NewStudyGroupService(groupRepo *postgres.StudyGroupRepository) *StudyGroupService {
	return &StudyGroupService{groupRepo: groupRepo}
}

// CreateGroupRequest represents a request to create a study group
type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic"`
	MaxMembers  int    `json:"maxMembers"`
}

// Create creates a new study group
func (s *StudyGroupService) Create(ctx context.Context, userID uuid.UUID, req *CreateGroupRequest) (*domain.StudyGroup, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	// Default max members if not specified
	maxMembers := req.MaxMembers
	if maxMembers <= 0 {
		maxMembers = 20
	}

	group := domain.NewStudyGroup(req.Name, req.Description, req.IsPublic, maxMembers, userID)
	if err := s.groupRepo.Create(ctx, group); err != nil {
		return nil, fmt.Errorf("failed to create study group: %w", err)
	}

	return group, nil
}

// GetByID retrieves a study group by ID
func (s *StudyGroupService) GetByID(ctx context.Context, id uuid.UUID) (*domain.StudyGroup, error) {
	return s.groupRepo.FindByID(ctx, id)
}

// ListByUser retrieves all study groups a user is a member of
func (s *StudyGroupService) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.StudyGroup, error) {
	return s.groupRepo.FindByUserID(ctx, userID)
}

// ListPublic retrieves all public study groups for discovery
func (s *StudyGroupService) ListPublic(ctx context.Context, limit, offset int) ([]domain.StudyGroup, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	groups, err := s.groupRepo.ListPublic(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.groupRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// Join adds a user to a study group
func (s *StudyGroupService) Join(ctx context.Context, groupID, userID uuid.UUID) error {
	// Check if group exists
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil || group == nil {
		return fmt.Errorf("study group not found")
	}

	member := &domain.StudyGroupMember{
		GroupID:  groupID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now().UTC(),
	}

	return s.groupRepo.AddMember(ctx, member)
}

// Leave removes a user from a study group
func (s *StudyGroupService) Leave(ctx context.Context, groupID, userID uuid.UUID) error {
	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

// GetMembers retrieves all members of a study group
func (s *StudyGroupService) GetMembers(ctx context.Context, groupID uuid.UUID) ([]domain.StudyGroupMember, error) {
	return s.groupRepo.GetMembers(ctx, groupID)
}

// IsMember checks if a user is a member of a study group
func (s *StudyGroupService) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	return s.groupRepo.IsMember(ctx, groupID, userID)
}

// Delete removes a study group (only by owner)
func (s *StudyGroupService) Delete(ctx context.Context, id, ownerID uuid.UUID) error {
	return s.groupRepo.Delete(ctx, id, ownerID)
}

// GetMemberCount returns the number of members in a group
func (s *StudyGroupService) GetMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	return s.groupRepo.GetMemberCount(ctx, groupID)
}
