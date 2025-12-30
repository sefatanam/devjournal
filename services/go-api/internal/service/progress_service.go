package service

import (
	"context"
	"fmt"
	"time"

	"devjournal/internal/domain"
	"devjournal/internal/repository/postgres"

	"github.com/google/uuid"
)

// ProgressService handles learning progress business logic
type ProgressService struct {
	progressRepo *postgres.ProgressRepository
}

// NewProgressService creates a new progress service
func NewProgressService(progressRepo *postgres.ProgressRepository) *ProgressService {
	return &ProgressService{progressRepo: progressRepo}
}

// GetSummary retrieves the learning progress summary for a user
func (s *ProgressService) GetSummary(ctx context.Context, userID uuid.UUID) (*domain.ProgressSummary, error) {
	summary, err := s.progressRepo.GetSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get progress summary: %w", err)
	}
	return summary, nil
}

// GetTodayProgress retrieves today's progress for a user
func (s *ProgressService) GetTodayProgress(ctx context.Context, userID uuid.UUID) (*domain.LearningProgress, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	progress, err := s.progressRepo.FindByUserAndDate(ctx, userID, today)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's progress: %w", err)
	}

	if progress == nil {
		// Return empty progress for today
		progress = domain.NewLearningProgress(userID, today)
	}

	return progress, nil
}

// GetWeeklyProgress retrieves the last 7 days of progress
func (s *ProgressService) GetWeeklyProgress(ctx context.Context, userID uuid.UUID) ([]domain.LearningProgress, error) {
	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -6)

	progressList, err := s.progressRepo.FindByUserRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly progress: %w", err)
	}

	// Return empty slice instead of nil to avoid JSON null
	if progressList == nil {
		return []domain.LearningProgress{}, nil
	}

	return progressList, nil
}

// GetMonthlyProgress retrieves the last 30 days of progress
func (s *ProgressService) GetMonthlyProgress(ctx context.Context, userID uuid.UUID) ([]domain.LearningProgress, error) {
	endDate := time.Now().UTC().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, 0, -29)

	progressList, err := s.progressRepo.FindByUserRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly progress: %w", err)
	}

	// Return empty slice instead of nil to avoid JSON null
	if progressList == nil {
		return []domain.LearningProgress{}, nil
	}

	return progressList, nil
}

// RecordJournalEntry records that a journal entry was created
func (s *ProgressService) RecordJournalEntry(ctx context.Context, userID uuid.UUID) error {
	if err := s.progressRepo.IncrementEntries(ctx, userID); err != nil {
		return fmt.Errorf("failed to record journal entry: %w", err)
	}

	// Update streak
	if err := s.updateStreak(ctx, userID); err != nil {
		return fmt.Errorf("failed to update streak: %w", err)
	}

	return nil
}

// RecordSnippet records that a snippet was created
func (s *ProgressService) RecordSnippet(ctx context.Context, userID uuid.UUID) error {
	if err := s.progressRepo.IncrementSnippets(ctx, userID); err != nil {
		return fmt.Errorf("failed to record snippet: %w", err)
	}

	// Update streak
	if err := s.updateStreak(ctx, userID); err != nil {
		return fmt.Errorf("failed to update streak: %w", err)
	}

	return nil
}

// updateStreak calculates and updates the current streak
func (s *ProgressService) updateStreak(ctx context.Context, userID uuid.UUID) error {
	streak, err := s.progressRepo.CalculateStreak(ctx, userID)
	if err != nil {
		return err
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	progress, err := s.progressRepo.FindByUserAndDate(ctx, userID, today)
	if err != nil {
		return err
	}

	if progress != nil {
		progress.StreakDays = streak
		return s.progressRepo.Upsert(ctx, progress)
	}

	return nil
}

// GetCurrentStreak returns the current learning streak
func (s *ProgressService) GetCurrentStreak(ctx context.Context, userID uuid.UUID) (int, error) {
	streak, err := s.progressRepo.CalculateStreak(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get current streak: %w", err)
	}
	return streak, nil
}
