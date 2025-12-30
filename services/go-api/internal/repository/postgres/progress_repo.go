package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"devjournal/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProgressRepository handles learning progress data persistence with raw SQL
type ProgressRepository struct {
	pool *pgxpool.Pool
}

// NewProgressRepository creates a new progress repository
func NewProgressRepository(pool *pgxpool.Pool) *ProgressRepository {
	return &ProgressRepository{pool: pool}
}

// Upsert creates or updates a progress record for a specific date
func (r *ProgressRepository) Upsert(ctx context.Context, progress *domain.LearningProgress) error {
	query := `
		INSERT INTO learning_progress (id, user_id, date, entries_count, snippets_count, streak_days, total_learning_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, date)
		DO UPDATE SET
			entries_count = $4,
			snippets_count = $5,
			streak_days = $6,
			total_learning_time = $7
	`
	_, err := r.pool.Exec(ctx, query,
		progress.ID,
		progress.UserID,
		progress.Date,
		progress.EntriesCount,
		progress.SnippetsCount,
		progress.StreakDays,
		progress.TotalLearningTime,
		progress.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert progress: %w", err)
	}
	return nil
}

// FindByUserAndDate retrieves progress for a specific user and date
func (r *ProgressRepository) FindByUserAndDate(ctx context.Context, userID uuid.UUID, date time.Time) (*domain.LearningProgress, error) {
	query := `
		SELECT id, user_id, date, entries_count, snippets_count, streak_days, total_learning_time, created_at
		FROM learning_progress
		WHERE user_id = $1 AND date = $2
	`
	row := r.pool.QueryRow(ctx, query, userID, date.Truncate(24*time.Hour))

	var progress domain.LearningProgress
	err := row.Scan(
		&progress.ID,
		&progress.UserID,
		&progress.Date,
		&progress.EntriesCount,
		&progress.SnippetsCount,
		&progress.StreakDays,
		&progress.TotalLearningTime,
		&progress.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find progress: %w", err)
	}
	return &progress, nil
}

// FindByUserRange retrieves progress records within a date range
func (r *ProgressRepository) FindByUserRange(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]domain.LearningProgress, error) {
	query := `
		SELECT id, user_id, date, entries_count, snippets_count, streak_days, total_learning_time, created_at
		FROM learning_progress
		WHERE user_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC
	`
	rows, err := r.pool.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to find progress range: %w", err)
	}
	defer rows.Close()

	var progressList []domain.LearningProgress
	for rows.Next() {
		var progress domain.LearningProgress
		err := rows.Scan(
			&progress.ID,
			&progress.UserID,
			&progress.Date,
			&progress.EntriesCount,
			&progress.SnippetsCount,
			&progress.StreakDays,
			&progress.TotalLearningTime,
			&progress.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan progress: %w", err)
		}
		progressList = append(progressList, progress)
	}

	return progressList, nil
}

// CalculateStreak calculates the current streak for a user
func (r *ProgressRepository) CalculateStreak(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		WITH RECURSIVE streak AS (
			SELECT date, 1 as streak_count
			FROM learning_progress
			WHERE user_id = $1 AND date = CURRENT_DATE AND (entries_count > 0 OR snippets_count > 0)

			UNION ALL

			SELECT lp.date, s.streak_count + 1
			FROM learning_progress lp
			JOIN streak s ON lp.date = s.date - INTERVAL '1 day'
			WHERE lp.user_id = $1 AND (lp.entries_count > 0 OR lp.snippets_count > 0)
		)
		SELECT COALESCE(MAX(streak_count), 0) FROM streak
	`
	var streak int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&streak)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate streak: %w", err)
	}
	return streak, nil
}

// GetSummary retrieves a summary of learning progress for a user
func (r *ProgressRepository) GetSummary(ctx context.Context, userID uuid.UUID) (*domain.ProgressSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(entries_count), 0) as total_entries,
			COALESCE(SUM(snippets_count), 0) as total_snippets,
			COALESCE(SUM(total_learning_time), 0) as total_time,
			COALESCE(MAX(streak_days), 0) as longest_streak
		FROM learning_progress
		WHERE user_id = $1
	`
	var summary domain.ProgressSummary
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&summary.TotalEntries,
		&summary.TotalSnippets,
		&summary.TotalLearningTime,
		&summary.LongestStreak,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	// Get current streak
	currentStreak, err := r.CalculateStreak(ctx, userID)
	if err != nil {
		return nil, err
	}
	summary.CurrentStreak = currentStreak

	// Get this week's entries
	weekQuery := `
		SELECT COALESCE(SUM(entries_count), 0)
		FROM learning_progress
		WHERE user_id = $1 AND date >= DATE_TRUNC('week', CURRENT_DATE)
	`
	err = r.pool.QueryRow(ctx, weekQuery, userID).Scan(&summary.ThisWeekEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly entries: %w", err)
	}

	// Get this month's entries
	monthQuery := `
		SELECT COALESCE(SUM(entries_count), 0)
		FROM learning_progress
		WHERE user_id = $1 AND date >= DATE_TRUNC('month', CURRENT_DATE)
	`
	err = r.pool.QueryRow(ctx, monthQuery, userID).Scan(&summary.ThisMonthEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly entries: %w", err)
	}

	return &summary, nil
}

// IncrementEntries increments the entry count for today
func (r *ProgressRepository) IncrementEntries(ctx context.Context, userID uuid.UUID) error {
	query := `
		INSERT INTO learning_progress (id, user_id, date, entries_count, created_at)
		VALUES ($1, $2, CURRENT_DATE, 1, NOW())
		ON CONFLICT (user_id, date)
		DO UPDATE SET entries_count = learning_progress.entries_count + 1
	`
	_, err := r.pool.Exec(ctx, query, uuid.New(), userID)
	if err != nil {
		return fmt.Errorf("failed to increment entries: %w", err)
	}
	return nil
}

// IncrementSnippets increments the snippet count for today
func (r *ProgressRepository) IncrementSnippets(ctx context.Context, userID uuid.UUID) error {
	query := `
		INSERT INTO learning_progress (id, user_id, date, snippets_count, created_at)
		VALUES ($1, $2, CURRENT_DATE, 1, NOW())
		ON CONFLICT (user_id, date)
		DO UPDATE SET snippets_count = learning_progress.snippets_count + 1
	`
	_, err := r.pool.Exec(ctx, query, uuid.New(), userID)
	if err != nil {
		return fmt.Errorf("failed to increment snippets: %w", err)
	}
	return nil
}
