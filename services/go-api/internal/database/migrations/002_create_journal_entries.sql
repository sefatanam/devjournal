-- Migration: Create journal_entries table
-- Description: Learning journal entries with mood tracking and tags

-- Up Migration
CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    mood VARCHAR(50),
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for user's entries (most common query)
CREATE INDEX IF NOT EXISTS idx_journal_entries_user_id ON journal_entries(user_id);

-- Index for chronological listing
CREATE INDEX IF NOT EXISTS idx_journal_entries_created_at ON journal_entries(created_at DESC);

-- Index for filtering by mood
CREATE INDEX IF NOT EXISTS idx_journal_entries_mood ON journal_entries(user_id, mood);

-- GIN index for array-based tag queries
CREATE INDEX IF NOT EXISTS idx_journal_entries_tags ON journal_entries USING GIN(tags);

-- Full-text search index
CREATE INDEX IF NOT EXISTS idx_journal_entries_search ON journal_entries
    USING GIN(to_tsvector('english', title || ' ' || content));

-- Auto-update trigger
CREATE TRIGGER update_journal_entries_updated_at
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Down Migration (commented out for safety)
-- DROP TRIGGER IF EXISTS update_journal_entries_updated_at ON journal_entries;
-- DROP TABLE IF EXISTS journal_entries;
