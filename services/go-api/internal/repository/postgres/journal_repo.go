package postgres

import (
	"context"
	"errors"
	"fmt"

	"devjournal/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// JournalRepository handles journal entry data persistence with raw SQL
type JournalRepository struct {
	pool *pgxpool.Pool
}

// NewJournalRepository creates a new journal repository
func NewJournalRepository(pool *pgxpool.Pool) *JournalRepository {
	return &JournalRepository{pool: pool}
}

// Create inserts a new journal entry
func (r *JournalRepository) Create(ctx context.Context, entry *domain.JournalEntry) error {
	query := `
		INSERT INTO journal_entries (id, user_id, title, content, mood, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		entry.ID,
		entry.UserID,
		entry.Title,
		entry.Content,
		entry.Mood,
		entry.Tags,
		entry.CreatedAt,
		entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}
	return nil
}

// FindByID retrieves a journal entry by ID
func (r *JournalRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.JournalEntry, error) {
	query := `
		SELECT id, user_id, title, content, mood, tags, created_at, updated_at
		FROM journal_entries
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)

	var entry domain.JournalEntry
	err := row.Scan(
		&entry.ID,
		&entry.UserID,
		&entry.Title,
		&entry.Content,
		&entry.Mood,
		&entry.Tags,
		&entry.CreatedAt,
		&entry.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find journal entry: %w", err)
	}
	return &entry, nil
}

// FindByUserID retrieves all journal entries for a user with pagination
func (r *JournalRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.JournalEntry, error) {
	query := `
		SELECT id, user_id, title, content, mood, tags, created_at, updated_at
		FROM journal_entries
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find journal entries: %w", err)
	}
	defer rows.Close()

	var entries []domain.JournalEntry
	for rows.Next() {
		var entry domain.JournalEntry
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Title,
			&entry.Content,
			&entry.Mood,
			&entry.Tags,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan journal entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating journal entries: %w", err)
	}

	return entries, nil
}

// FindByMood retrieves journal entries filtered by mood
func (r *JournalRepository) FindByMood(ctx context.Context, userID uuid.UUID, mood string, limit, offset int) ([]domain.JournalEntry, error) {
	query := `
		SELECT id, user_id, title, content, mood, tags, created_at, updated_at
		FROM journal_entries
		WHERE user_id = $1 AND mood = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, userID, mood, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find journal entries by mood: %w", err)
	}
	defer rows.Close()

	var entries []domain.JournalEntry
	for rows.Next() {
		var entry domain.JournalEntry
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Title,
			&entry.Content,
			&entry.Mood,
			&entry.Tags,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan journal entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Search searches journal entries by title or content
func (r *JournalRepository) Search(ctx context.Context, userID uuid.UUID, searchTerm string, limit, offset int) ([]domain.JournalEntry, error) {
	query := `
		SELECT id, user_id, title, content, mood, tags, created_at, updated_at
		FROM journal_entries
		WHERE user_id = $1
		  AND (title ILIKE $2 OR content ILIKE $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	searchPattern := "%" + searchTerm + "%"
	rows, err := r.pool.Query(ctx, query, userID, searchPattern, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search journal entries: %w", err)
	}
	defer rows.Close()

	var entries []domain.JournalEntry
	for rows.Next() {
		var entry domain.JournalEntry
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Title,
			&entry.Content,
			&entry.Mood,
			&entry.Tags,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan journal entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Update updates an existing journal entry
func (r *JournalRepository) Update(ctx context.Context, entry *domain.JournalEntry) error {
	query := `
		UPDATE journal_entries
		SET title = $2, content = $3, mood = $4, tags = $5, updated_at = $6
		WHERE id = $1 AND user_id = $7
	`
	result, err := r.pool.Exec(ctx, query,
		entry.ID,
		entry.Title,
		entry.Content,
		entry.Mood,
		entry.Tags,
		entry.UpdatedAt,
		entry.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update journal entry: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("journal entry not found or unauthorized")
	}
	return nil
}

// Delete removes a journal entry
func (r *JournalRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `DELETE FROM journal_entries WHERE id = $1 AND user_id = $2`
	result, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete journal entry: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("journal entry not found or unauthorized")
	}
	return nil
}

// Count returns the total number of entries for a user
func (r *JournalRepository) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM journal_entries WHERE user_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count journal entries: %w", err)
	}
	return count, nil
}
