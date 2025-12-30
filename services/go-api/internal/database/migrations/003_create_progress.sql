-- Migration: Create learning_progress table
-- Description: Daily learning progress tracking and streak calculation

-- Up Migration
CREATE TABLE IF NOT EXISTS learning_progress (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    entries_count INTEGER DEFAULT 0,
    snippets_count INTEGER DEFAULT 0,
    streak_days INTEGER DEFAULT 0,
    total_learning_time INTEGER DEFAULT 0, -- in minutes
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date)
);

-- Index for user's progress by date
CREATE INDEX IF NOT EXISTS idx_progress_user_date ON learning_progress(user_id, date DESC);

-- Index for streak calculation (recent entries)
CREATE INDEX IF NOT EXISTS idx_progress_user_recent ON learning_progress(user_id, date)
    WHERE entries_count > 0 OR snippets_count > 0;

-- Down Migration (commented out for safety)
-- DROP TABLE IF EXISTS learning_progress;
