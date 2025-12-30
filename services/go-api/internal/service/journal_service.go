package service

import (
	"context"
	"fmt"
	"time"

	"devjournal/internal/domain"
	"devjournal/internal/repository/postgres"

	"github.com/google/uuid"
)

// JournalService handles journal entry business logic
type JournalService struct {
	journalRepo *postgres.JournalRepository
}

// NewJournalService creates a new journal service
func NewJournalService(journalRepo *postgres.JournalRepository) *JournalService {
	return &JournalService{journalRepo: journalRepo}
}

// Create creates a new journal entry
func (s *JournalService) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateJournalEntryRequest) (*domain.JournalEntry, error) {
	entry := domain.NewJournalEntry(userID, req.Title, req.Content, req.Mood, req.Tags)

	if err := s.journalRepo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create journal entry: %w", err)
	}

	return entry, nil
}

// GetByID retrieves a journal entry by ID
func (s *JournalService) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.JournalEntry, error) {
	entry, err := s.journalRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find journal entry: %w", err)
	}
	if entry == nil {
		return nil, nil
	}
	// Verify ownership
	if entry.UserID != userID {
		return nil, nil
	}
	return entry, nil
}

// List retrieves all journal entries for a user
func (s *JournalService) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.JournalEntry, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	entries, err := s.journalRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list journal entries: %w", err)
	}

	total, err := s.journalRepo.Count(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count journal entries: %w", err)
	}

	return entries, total, nil
}

// ListByMood retrieves journal entries filtered by mood
func (s *JournalService) ListByMood(ctx context.Context, userID uuid.UUID, mood string, limit, offset int) ([]domain.JournalEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	entries, err := s.journalRepo.FindByMood(ctx, userID, mood, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list journal entries by mood: %w", err)
	}

	return entries, nil
}

// Search searches journal entries by title or content
func (s *JournalService) Search(ctx context.Context, userID uuid.UUID, searchTerm string, limit, offset int) ([]domain.JournalEntry, error) {
	if limit <= 0 {
		limit = 20
	}

	entries, err := s.journalRepo.Search(ctx, userID, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search journal entries: %w", err)
	}

	return entries, nil
}

// Update updates an existing journal entry
func (s *JournalService) Update(ctx context.Context, id, userID uuid.UUID, req *domain.UpdateJournalEntryRequest) (*domain.JournalEntry, error) {
	// Verify entry exists and belongs to user
	existing, err := s.journalRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find journal entry: %w", err)
	}
	if existing == nil || existing.UserID != userID {
		return nil, fmt.Errorf("journal entry not found")
	}

	// Update fields
	existing.Title = req.Title
	existing.Content = req.Content
	existing.Mood = req.Mood
	existing.Tags = req.Tags
	existing.UpdatedAt = time.Now().UTC()

	if err := s.journalRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update journal entry: %w", err)
	}

	return existing, nil
}

// Delete removes a journal entry
func (s *JournalService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if err := s.journalRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete journal entry: %w", err)
	}
	return nil
}
