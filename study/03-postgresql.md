# PostgreSQL Database

## Overview

DevJournal uses PostgreSQL for relational data that benefits from:
- ACID transactions
- Referential integrity (foreign keys)
- Complex queries (joins, aggregations)
- Structured, predictable schema

## What's Stored in PostgreSQL?

| Table | Purpose | Why PostgreSQL? |
|-------|---------|-----------------|
| `users` | User accounts | Core entity, needs integrity |
| `journal_entries` | Dev logs | Structured with user FK |
| `learning_progress` | Daily stats | Aggregations, time-series |
| `study_groups` | Group metadata | Relationships between users |
| `study_group_members` | Memberships | Many-to-many relationship |

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for email lookups (login)
CREATE INDEX idx_users_email ON users(email);
```

### Journal Entries Table

```sql
CREATE TABLE journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    mood VARCHAR(50),
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for user's entries (most common query)
CREATE INDEX idx_journal_entries_user_id ON journal_entries(user_id);

-- Index for date-based queries
CREATE INDEX idx_journal_entries_created_at ON journal_entries(created_at DESC);

-- Index for mood filtering
CREATE INDEX idx_journal_entries_mood ON journal_entries(mood) WHERE mood IS NOT NULL;

-- Full-text search index
CREATE INDEX idx_journal_entries_search ON journal_entries
    USING GIN (to_tsvector('english', title || ' ' || content));
```

### Learning Progress Table

```sql
CREATE TABLE learning_progress (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    entries_count INTEGER DEFAULT 0,
    snippets_created INTEGER DEFAULT 0,
    study_minutes INTEGER DEFAULT 0,
    streak_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, date)
);

-- Index for user progress queries
CREATE INDEX idx_progress_user_date ON learning_progress(user_id, date DESC);
```

### Study Groups Tables

```sql
CREATE TABLE study_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN DEFAULT false,
    max_members INTEGER DEFAULT 50,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for public group discovery
CREATE INDEX idx_groups_public ON study_groups(is_public) WHERE is_public = true;

-- Junction table for many-to-many
CREATE TABLE study_group_members (
    group_id UUID NOT NULL REFERENCES study_groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

-- Index for user's groups
CREATE INDEX idx_group_members_user ON study_group_members(user_id);
```

## Go Database Connection

### Connection Pool with pgx

```go
// services/go-api/internal/database/postgres.go

package database

import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
    Pool *pgxpool.Pool
}

func NewPostgresDB(ctx context.Context, databaseURL string) (*PostgresDB, error) {
    // Parse connection config
    config, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to parse database URL: %w", err)
    }

    // Configure pool settings
    config.MinConns = 5                        // Minimum connections to keep open
    config.MaxConns = 25                       // Maximum connections allowed
    config.MaxConnLifetime = 15 * time.Minute  // Connection max lifetime
    config.MaxConnIdleTime = 5 * time.Minute   // Idle connection timeout
    config.HealthCheckPeriod = 1 * time.Minute // Health check interval

    // Create connection pool
    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create pool: %w", err)
    }

    // Test connection
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return &PostgresDB{Pool: pool}, nil
}

func (db *PostgresDB) Close() {
    db.Pool.Close()
}
```

## Repository Pattern

### Interface Definition

```go
// services/go-api/internal/repository/interfaces.go

package repository

import (
    "context"
    "github.com/devjournal/internal/domain"
)

type JournalRepository interface {
    Create(ctx context.Context, entry *domain.JournalEntry) (*domain.JournalEntry, error)
    GetByID(ctx context.Context, id string) (*domain.JournalEntry, error)
    List(ctx context.Context, userID string, filter domain.JournalFilter) ([]*domain.JournalEntry, int, error)
    Update(ctx context.Context, entry *domain.JournalEntry) (*domain.JournalEntry, error)
    Delete(ctx context.Context, id string) error
    Search(ctx context.Context, userID string, query string) ([]*domain.JournalEntry, error)
}
```

### PostgreSQL Implementation

```go
// services/go-api/internal/repository/postgres/journal_repo.go

package postgres

import (
    "context"
    "errors"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type JournalRepo struct {
    pool *pgxpool.Pool
}

func NewJournalRepo(pool *pgxpool.Pool) *JournalRepo {
    return &JournalRepo{pool: pool}
}

// Create inserts a new journal entry
func (r *JournalRepo) Create(ctx context.Context, entry *domain.JournalEntry) (*domain.JournalEntry, error) {
    query := `
        INSERT INTO journal_entries (user_id, title, content, mood, tags)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
    `

    err := r.pool.QueryRow(ctx, query,
        entry.UserID,
        entry.Title,
        entry.Content,
        entry.Mood,
        entry.Tags,
    ).Scan(&entry.ID, &entry.CreatedAt, &entry.UpdatedAt)

    if err != nil {
        return nil, fmt.Errorf("failed to create entry: %w", err)
    }

    return entry, nil
}

// GetByID retrieves a single entry
func (r *JournalRepo) GetByID(ctx context.Context, id string) (*domain.JournalEntry, error) {
    query := `
        SELECT id, user_id, title, content, mood, tags, created_at, updated_at
        FROM journal_entries
        WHERE id = $1
    `

    entry := &domain.JournalEntry{}
    err := r.pool.QueryRow(ctx, query, id).Scan(
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
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get entry: %w", err)
    }

    return entry, nil
}

// List retrieves paginated entries with filters
func (r *JournalRepo) List(
    ctx context.Context,
    userID string,
    filter domain.JournalFilter,
) ([]*domain.JournalEntry, int, error) {
    // Build dynamic query based on filters
    baseQuery := `
        SELECT id, user_id, title, content, mood, tags, created_at, updated_at
        FROM journal_entries
        WHERE user_id = $1
    `
    countQuery := `
        SELECT COUNT(*)
        FROM journal_entries
        WHERE user_id = $1
    `

    args := []interface{}{userID}
    argIndex := 2

    // Add mood filter if specified
    if filter.Mood != "" {
        moodCondition := fmt.Sprintf(" AND mood = $%d", argIndex)
        baseQuery += moodCondition
        countQuery += moodCondition
        args = append(args, filter.Mood)
        argIndex++
    }

    // Add search filter if specified
    if filter.Search != "" {
        searchCondition := fmt.Sprintf(
            " AND to_tsvector('english', title || ' ' || content) @@ plainto_tsquery('english', $%d)",
            argIndex,
        )
        baseQuery += searchCondition
        countQuery += searchCondition
        args = append(args, filter.Search)
        argIndex++
    }

    // Get total count
    var total int
    err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to count entries: %w", err)
    }

    // Add pagination and ordering
    baseQuery += " ORDER BY created_at DESC"
    baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)

    offset := (filter.Page - 1) * filter.PageSize
    args = append(args, filter.PageSize, offset)

    // Execute query
    rows, err := r.pool.Query(ctx, baseQuery, args...)
    if err != nil {
        return nil, 0, fmt.Errorf("failed to list entries: %w", err)
    }
    defer rows.Close()

    // Scan results
    entries := make([]*domain.JournalEntry, 0)
    for rows.Next() {
        entry := &domain.JournalEntry{}
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
            return nil, 0, fmt.Errorf("failed to scan entry: %w", err)
        }
        entries = append(entries, entry)
    }

    return entries, total, nil
}

// Update modifies an existing entry
func (r *JournalRepo) Update(ctx context.Context, entry *domain.JournalEntry) (*domain.JournalEntry, error) {
    query := `
        UPDATE journal_entries
        SET title = $1, content = $2, mood = $3, tags = $4, updated_at = NOW()
        WHERE id = $5 AND user_id = $6
        RETURNING updated_at
    `

    err := r.pool.QueryRow(ctx, query,
        entry.Title,
        entry.Content,
        entry.Mood,
        entry.Tags,
        entry.ID,
        entry.UserID,
    ).Scan(&entry.UpdatedAt)

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to update entry: %w", err)
    }

    return entry, nil
}

// Delete removes an entry
func (r *JournalRepo) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM journal_entries WHERE id = $1`

    result, err := r.pool.Exec(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete entry: %w", err)
    }

    if result.RowsAffected() == 0 {
        return ErrNotFound
    }

    return nil
}

// Search performs full-text search
func (r *JournalRepo) Search(
    ctx context.Context,
    userID string,
    query string,
) ([]*domain.JournalEntry, error) {
    sql := `
        SELECT id, user_id, title, content, mood, tags, created_at, updated_at,
               ts_rank(to_tsvector('english', title || ' ' || content),
                       plainto_tsquery('english', $2)) as rank
        FROM journal_entries
        WHERE user_id = $1
          AND to_tsvector('english', title || ' ' || content) @@ plainto_tsquery('english', $2)
        ORDER BY rank DESC
        LIMIT 20
    `

    rows, err := r.pool.Query(ctx, sql, userID, query)
    if err != nil {
        return nil, fmt.Errorf("failed to search entries: %w", err)
    }
    defer rows.Close()

    entries := make([]*domain.JournalEntry, 0)
    for rows.Next() {
        entry := &domain.JournalEntry{}
        var rank float64
        err := rows.Scan(
            &entry.ID,
            &entry.UserID,
            &entry.Title,
            &entry.Content,
            &entry.Mood,
            &entry.Tags,
            &entry.CreatedAt,
            &entry.UpdatedAt,
            &rank,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan entry: %w", err)
        }
        entries = append(entries, entry)
    }

    return entries, nil
}
```

## Transaction Example

```go
// Using transactions for atomic operations
func (r *ProgressRepo) RecordActivity(ctx context.Context, userID string, entryCreated bool) error {
    tx, err := r.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx) // Rollback if not committed

    today := time.Now().Format("2006-01-02")

    // Upsert progress record
    query := `
        INSERT INTO learning_progress (user_id, date, entries_count)
        VALUES ($1, $2, 1)
        ON CONFLICT (user_id, date)
        DO UPDATE SET entries_count = learning_progress.entries_count + 1
    `

    _, err = tx.Exec(ctx, query, userID, today)
    if err != nil {
        return fmt.Errorf("failed to update progress: %w", err)
    }

    // Update streak
    streakQuery := `
        UPDATE learning_progress
        SET streak_count = (
            SELECT COUNT(DISTINCT date)
            FROM learning_progress
            WHERE user_id = $1
              AND date >= CURRENT_DATE - INTERVAL '30 days'
              AND (entries_count > 0 OR snippets_created > 0)
        )
        WHERE user_id = $1 AND date = $2
    `

    _, err = tx.Exec(ctx, streakQuery, userID, today)
    if err != nil {
        return fmt.Errorf("failed to update streak: %w", err)
    }

    // Commit transaction
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}
```

## PostgreSQL Features Used

### 1. UUID Primary Keys
```sql
id UUID PRIMARY KEY DEFAULT gen_random_uuid()
```
- Globally unique identifiers
- No sequence conflicts in distributed systems

### 2. Array Types
```sql
tags TEXT[] DEFAULT '{}'
```
- Native array support for tags
- Can query with `@>` (contains) operator

### 3. Full-Text Search
```sql
-- Creating the index
CREATE INDEX idx_journal_entries_search ON journal_entries
    USING GIN (to_tsvector('english', title || ' ' || content));

-- Querying with ranking
SELECT *, ts_rank(to_tsvector('english', title || ' ' || content),
                  plainto_tsquery('english', 'angular signals')) as rank
FROM journal_entries
WHERE to_tsvector('english', title || ' ' || content)
      @@ plainto_tsquery('english', 'angular signals')
ORDER BY rank DESC;
```

### 4. Timestamp with Time Zone
```sql
created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
```
- Proper timezone handling
- Automatic conversion to UTC

### 5. Cascading Deletes
```sql
user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
```
- Automatic cleanup of related records

### 6. Upsert (INSERT ON CONFLICT)
```sql
INSERT INTO learning_progress (user_id, date, entries_count)
VALUES ($1, $2, 1)
ON CONFLICT (user_id, date)
DO UPDATE SET entries_count = learning_progress.entries_count + 1
```

## Connection String Format

```
postgresql://username:password@host:port/database?sslmode=disable

# Example
postgresql://devjournal:devpass@localhost:5432/devjournal?sslmode=disable
```

## Docker Setup

```yaml
# docker/docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: devjournal
      POSTGRES_PASSWORD: devpass
      POSTGRES_DB: devjournal
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U devjournal"]
      interval: 5s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
```

## Key Takeaways

1. **pgx over database/sql** - Better performance and PostgreSQL-specific features
2. **Connection pooling** - Essential for production (pgxpool)
3. **Parameterized queries** - Always use $1, $2 placeholders to prevent SQL injection
4. **Repository pattern** - Clean separation between data access and business logic
5. **Transactions** - Use for multi-step operations that must be atomic
6. **Indexes** - Create for frequently queried columns
7. **Full-text search** - Use GIN indexes for text search instead of LIKE

## Next Steps

- [MongoDB Database](./04-mongodb.md) - Document storage for snippets
- [Authentication & JWT](./07-authentication.md) - User management
