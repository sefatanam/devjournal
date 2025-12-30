package postgres

import (
	"context"
	"fmt"

	"devjournal/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StudyGroupRepository handles study group database operations
type StudyGroupRepository struct {
	pool *pgxpool.Pool
}

// NewStudyGroupRepository creates a new study group repository
func NewStudyGroupRepository(pool *pgxpool.Pool) *StudyGroupRepository {
	return &StudyGroupRepository{pool: pool}
}

// Create creates a new study group and adds the creator as owner
func (r *StudyGroupRepository) Create(ctx context.Context, group *domain.StudyGroup) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert study group
	_, err = tx.Exec(ctx, `
		INSERT INTO study_groups (id, name, description, is_public, max_members, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, group.ID, group.Name, group.Description, group.IsPublic, group.MaxMembers, group.CreatedBy, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert study group: %w", err)
	}

	// Add creator as owner
	_, err = tx.Exec(ctx, `
		INSERT INTO study_group_members (group_id, user_id, role, joined_at)
		VALUES ($1, $2, 'owner', $3)
	`, group.ID, group.CreatedBy, group.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to add creator as owner: %w", err)
	}

	return tx.Commit(ctx)
}

// FindByID retrieves a study group by ID
func (r *StudyGroupRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.StudyGroup, error) {
	var group domain.StudyGroup
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, description, is_public, max_members, created_by, created_at, updated_at
		FROM study_groups
		WHERE id = $1
	`, id).Scan(&group.ID, &group.Name, &group.Description, &group.IsPublic, &group.MaxMembers, &group.CreatedBy, &group.CreatedAt, &group.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// FindByUserID retrieves all study groups a user is a member of
func (r *StudyGroupRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]domain.StudyGroup, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sg.id, sg.name, sg.description, sg.is_public, sg.max_members, sg.created_by, sg.created_at, sg.updated_at
		FROM study_groups sg
		JOIN study_group_members sgm ON sg.id = sgm.group_id
		WHERE sgm.user_id = $1
		ORDER BY sg.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query study groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.StudyGroup
	for rows.Next() {
		var group domain.StudyGroup
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.IsPublic, &group.MaxMembers, &group.CreatedBy, &group.CreatedAt, &group.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan study group: %w", err)
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// ListPublic retrieves all public study groups (for discovery)
func (r *StudyGroupRepository) ListPublic(ctx context.Context, limit, offset int) ([]domain.StudyGroup, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description, is_public, max_members, created_by, created_at, updated_at
		FROM study_groups
		WHERE is_public = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query study groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.StudyGroup
	for rows.Next() {
		var group domain.StudyGroup
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.IsPublic, &group.MaxMembers, &group.CreatedBy, &group.CreatedAt, &group.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan study group: %w", err)
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// AddMember adds a user to a study group
func (r *StudyGroupRepository) AddMember(ctx context.Context, member *domain.StudyGroupMember) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO study_group_members (group_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, member.GroupID, member.UserID, member.Role, member.JoinedAt)
	return err
}

// RemoveMember removes a user from a study group
func (r *StudyGroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM study_group_members
		WHERE group_id = $1 AND user_id = $2 AND role != 'owner'
	`, groupID, userID)
	return err
}

// GetMembers retrieves all members of a study group with display names
func (r *StudyGroupRepository) GetMembers(ctx context.Context, groupID uuid.UUID) ([]domain.StudyGroupMember, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT sgm.group_id, sgm.user_id, u.display_name, sgm.role, sgm.joined_at
		FROM study_group_members sgm
		JOIN users u ON sgm.user_id = u.id
		WHERE sgm.group_id = $1
		ORDER BY sgm.joined_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	var members []domain.StudyGroupMember
	for rows.Next() {
		var member domain.StudyGroupMember
		if err := rows.Scan(&member.GroupID, &member.UserID, &member.DisplayName, &member.Role, &member.JoinedAt); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

// IsMember checks if a user is a member of a study group
func (r *StudyGroupRepository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM study_group_members
			WHERE group_id = $1 AND user_id = $2
		)
	`, groupID, userID).Scan(&exists)
	return exists, err
}

// Delete removes a study group (only by owner)
func (r *StudyGroupRepository) Delete(ctx context.Context, id, ownerID uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		DELETE FROM study_groups
		WHERE id = $1 AND created_by = $2
	`, id, ownerID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("study group not found or not authorized")
	}
	return nil
}

// Count returns the total number of study groups
func (r *StudyGroupRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM study_groups`).Scan(&count)
	return count, err
}

// GetMemberCount returns the number of members in a group
func (r *StudyGroupRepository) GetMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM study_group_members WHERE group_id = $1
	`, groupID).Scan(&count)
	return count, err
}
