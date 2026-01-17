-- Would Watch Backend - Database Schema
-- This schema should be run against your Supabase PostgreSQL database

-- ============================================================================
-- TYPES
-- ============================================================================

-- Create enum for media types
CREATE TYPE media_type AS ENUM ('movie', 'tv');

-- Create enum for session status
CREATE TYPE session_status AS ENUM ('active', 'completed');

-- ============================================================================
-- TABLES
-- ============================================================================

-- Media Items Table
-- Stores movie and TV show information from TMDB
CREATE TABLE IF NOT EXISTS media_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tmdb_id INTEGER NOT NULL,
    media_type media_type NOT NULL,
    title TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Ensure we don't duplicate TMDB entries
    CONSTRAINT unique_tmdb_media UNIQUE (tmdb_id, media_type)
);

-- Watch Sessions Table
-- Stores group watch sessions for collaborative movie selection
CREATE TABLE IF NOT EXISTS watch_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL,
    status session_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    -- Foreign key to Supabase auth.users table
    CONSTRAINT fk_creator FOREIGN KEY (creator_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Index for faster TMDB ID lookups
CREATE INDEX IF NOT EXISTS idx_media_items_tmdb_id
    ON media_items(tmdb_id);

-- Index for media type filtering
CREATE INDEX IF NOT EXISTS idx_media_items_type
    ON media_items(media_type);

-- Index for session status queries
CREATE INDEX IF NOT EXISTS idx_watch_sessions_status
    ON watch_sessions(status);

-- Index for user's sessions lookup
CREATE INDEX IF NOT EXISTS idx_watch_sessions_creator
    ON watch_sessions(creator_id);

-- Composite index for active sessions by user
CREATE INDEX IF NOT EXISTS idx_watch_sessions_creator_status
    ON watch_sessions(creator_id, status);

-- Index for recently created sessions
CREATE INDEX IF NOT EXISTS idx_watch_sessions_created
    ON watch_sessions(created_at DESC);

-- GIN index for efficient JSONB queries on metadata
CREATE INDEX IF NOT EXISTS idx_media_items_metadata
    ON media_items USING GIN (metadata);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for media_items
DROP TRIGGER IF EXISTS update_media_items_updated_at ON media_items;
CREATE TRIGGER update_media_items_updated_at
    BEFORE UPDATE ON media_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger for watch_sessions
DROP TRIGGER IF EXISTS update_watch_sessions_updated_at ON watch_sessions;
CREATE TRIGGER update_watch_sessions_updated_at
    BEFORE UPDATE ON watch_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to automatically set completed_at when status changes to completed
CREATE OR REPLACE FUNCTION set_completed_at()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'completed' AND OLD.status != 'completed' THEN
        NEW.completed_at = NOW();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to set completed_at timestamp
DROP TRIGGER IF EXISTS set_watch_session_completed_at ON watch_sessions;
CREATE TRIGGER set_watch_session_completed_at
    BEFORE UPDATE ON watch_sessions
    FOR EACH ROW
    EXECUTE FUNCTION set_completed_at();

-- ============================================================================
-- ROW LEVEL SECURITY (RLS)
-- ============================================================================

-- Enable RLS on tables
ALTER TABLE media_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE watch_sessions ENABLE ROW LEVEL SECURITY;

-- Media Items Policies
-- Allow authenticated users to read all media items
CREATE POLICY "Allow authenticated users to read media items"
    ON media_items
    FOR SELECT
    TO authenticated
    USING (true);

-- Allow authenticated users to insert media items
CREATE POLICY "Allow authenticated users to insert media items"
    ON media_items
    FOR INSERT
    TO authenticated
    WITH CHECK (true);

-- Watch Sessions Policies
-- Users can read sessions they created
CREATE POLICY "Users can read own sessions"
    ON watch_sessions
    FOR SELECT
    TO authenticated
    USING (auth.uid() = creator_id);

-- Users can create sessions
CREATE POLICY "Users can create sessions"
    ON watch_sessions
    FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = creator_id);

-- Users can update their own sessions
CREATE POLICY "Users can update own sessions"
    ON watch_sessions
    FOR UPDATE
    TO authenticated
    USING (auth.uid() = creator_id)
    WITH CHECK (auth.uid() = creator_id);

-- Users can delete their own sessions
CREATE POLICY "Users can delete own sessions"
    ON watch_sessions
    FOR DELETE
    TO authenticated
    USING (auth.uid() = creator_id);

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE media_items IS 'Stores movie and TV show metadata from TMDB API';
COMMENT ON COLUMN media_items.tmdb_id IS 'The Movie Database (TMDB) ID';
COMMENT ON COLUMN media_items.media_type IS 'Type of media: movie or tv';
COMMENT ON COLUMN media_items.metadata IS 'JSONB field storing poster_path, overview, release_date, ratings, etc.';

COMMENT ON TABLE watch_sessions IS 'Group watch sessions for collaborative movie selection';
COMMENT ON COLUMN watch_sessions.creator_id IS 'User ID of the session creator (references auth.users)';
COMMENT ON COLUMN watch_sessions.status IS 'Session status: active or completed';
COMMENT ON COLUMN watch_sessions.completed_at IS 'Timestamp when session was marked as completed';
