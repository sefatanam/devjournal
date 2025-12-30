-- Migration: Create study_groups and study_group_members tables
-- Description: Study group chat rooms for collaboration

-- Up Migration
CREATE TABLE IF NOT EXISTS study_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    max_members INTEGER DEFAULT 50,
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS study_group_members (
    group_id UUID NOT NULL REFERENCES study_groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member', -- owner, admin, member
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id)
);

-- Index for finding user's groups
CREATE INDEX IF NOT EXISTS idx_study_group_members_user ON study_group_members(user_id);

-- Index for listing group members
CREATE INDEX IF NOT EXISTS idx_study_group_members_group ON study_group_members(group_id);

-- Index for public groups discovery
CREATE INDEX IF NOT EXISTS idx_study_groups_public ON study_groups(is_public) WHERE is_public = true;

-- Down Migration (commented out for safety)
-- DROP TABLE IF EXISTS study_group_members;
-- DROP TABLE IF EXISTS study_groups;
