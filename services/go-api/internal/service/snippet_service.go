package service

import (
	"context"
	"fmt"
	"time"

	"devjournal/internal/domain"
	"devjournal/internal/repository/mongodb"
)

// SnippetService handles code snippet business logic
type SnippetService struct {
	snippetRepo *mongodb.SnippetRepository
}

// NewSnippetService creates a new snippet service
func NewSnippetService(snippetRepo *mongodb.SnippetRepository) *SnippetService {
	return &SnippetService{snippetRepo: snippetRepo}
}

// Create creates a new code snippet
func (s *SnippetService) Create(ctx context.Context, userID string, req *domain.CreateSnippetRequest) (*domain.Snippet, error) {
	snippet := domain.NewSnippet(
		userID,
		req.Title,
		req.Description,
		req.Code,
		req.Language,
		req.Tags,
		req.Metadata,
		req.IsPublic,
	)

	if err := s.snippetRepo.Create(ctx, snippet); err != nil {
		return nil, fmt.Errorf("failed to create snippet: %w", err)
	}

	return snippet, nil
}

// GetByID retrieves a snippet by ID
func (s *SnippetService) GetByID(ctx context.Context, id, userID string) (*domain.Snippet, error) {
	snippet, err := s.snippetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find snippet: %w", err)
	}
	if snippet == nil {
		return nil, nil
	}

	// Check access
	if snippet.UserID != userID && !snippet.IsPublic {
		return nil, nil
	}

	// Increment views if not owner
	if snippet.UserID != userID {
		s.snippetRepo.IncrementViews(ctx, id)
	}

	return snippet, nil
}

// List retrieves all snippets for a user
func (s *SnippetService) List(ctx context.Context, userID string, limit, offset int64) ([]domain.Snippet, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	snippets, err := s.snippetRepo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list snippets: %w", err)
	}

	total, err := s.snippetRepo.Count(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count snippets: %w", err)
	}

	return snippets, total, nil
}

// ListByTags retrieves snippets matching any of the given tags
func (s *SnippetService) ListByTags(ctx context.Context, userID string, tags []string, limit, offset int64) ([]domain.Snippet, error) {
	if limit <= 0 {
		limit = 20
	}

	snippets, err := s.snippetRepo.FindByTags(ctx, userID, tags, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list snippets by tags: %w", err)
	}

	return snippets, nil
}

// ListByLanguage retrieves snippets by programming language
func (s *SnippetService) ListByLanguage(ctx context.Context, userID, language string, limit, offset int64) ([]domain.Snippet, error) {
	if limit <= 0 {
		limit = 20
	}

	snippets, err := s.snippetRepo.FindByLanguage(ctx, userID, language, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list snippets by language: %w", err)
	}

	return snippets, nil
}

// Search performs full-text search on snippets
func (s *SnippetService) Search(ctx context.Context, userID, query string, limit, offset int64) ([]domain.Snippet, error) {
	if limit <= 0 {
		limit = 20
	}

	snippets, err := s.snippetRepo.Search(ctx, userID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search snippets: %w", err)
	}

	return snippets, nil
}

// Update updates an existing snippet
func (s *SnippetService) Update(ctx context.Context, id, userID string, req *domain.UpdateSnippetRequest) (*domain.Snippet, error) {
	// Verify snippet exists and belongs to user
	existing, err := s.snippetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find snippet: %w", err)
	}
	if existing == nil || existing.UserID != userID {
		return nil, fmt.Errorf("snippet not found")
	}

	// Update fields
	existing.Title = req.Title
	existing.Description = req.Description
	existing.Code = req.Code
	existing.Language = req.Language
	existing.Tags = req.Tags
	existing.Metadata = req.Metadata
	existing.IsPublic = req.IsPublic
	existing.UpdatedAt = time.Now().UTC()

	if err := s.snippetRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update snippet: %w", err)
	}

	return existing, nil
}

// Delete removes a snippet
func (s *SnippetService) Delete(ctx context.Context, id, userID string) error {
	if err := s.snippetRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("failed to delete snippet: %w", err)
	}
	return nil
}

// GetLanguageStats returns snippet counts grouped by language
func (s *SnippetService) GetLanguageStats(ctx context.Context, userID string) (map[string]int64, error) {
	stats, err := s.snippetRepo.GetLanguageStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get language stats: %w", err)
	}
	return stats, nil
}
