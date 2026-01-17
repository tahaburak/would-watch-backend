-- Would Watch Backend - Database Schema
-- This schema should be run against your Supabase PostgreSQL database

-- ============================================================================
-- TYPES
-- ============================================================================

-- Create enum for media types
CREATE TYPE media_type AS ENUM ('movie', 'tv');

-- Create enum for session status
CREATE TYPE session_status AS ENUM ('active', 'completed');

-- Create enum for vote types
CREATE TYPE vote_type AS ENUM ('yes', 'no', 'maybe');

-- Create enum for invite preferences
CREATE TYPE invite_preference AS ENUM ('everyone', 'following', 'none');

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

-- Watch Sessions Table (now Rooms)
-- Stores group watch sessions for collaborative movie selection
CREATE TABLE IF NOT EXISTS watch_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL,
    name TEXT,
    is_public BOOLEAN DEFAULT false,
    status session_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    -- Foreign key to Supabase auth.users table
    CONSTRAINT fk_creator FOREIGN KEY (creator_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE
);

-- Session Votes Table
-- Stores user votes for media items within watch sessions
CREATE TABLE IF NOT EXISTS session_votes (
    session_id UUID NOT NULL,
    user_id UUID NOT NULL,
    media_id UUID NOT NULL,
    vote vote_type NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- Foreign keys
    CONSTRAINT fk_session FOREIGN KEY (session_id)
        REFERENCES watch_sessions(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_media FOREIGN KEY (media_id)
        REFERENCES media_items(id)
        ON DELETE CASCADE,

    -- Unique constraint: one vote per user per media per session
    CONSTRAINT unique_user_media_session UNIQUE (session_id, user_id, media_id)
);

-- User Profiles Table
-- Stores user profile information and privacy settings
CREATE TABLE IF NOT EXISTS profiles (
    user_id UUID PRIMARY KEY,
    username TEXT UNIQUE,
    invite_preference invite_preference DEFAULT 'following',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Foreign key to Supabase auth.users table
    CONSTRAINT fk_profile_user FOREIGN KEY (user_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE
);

-- User Follows Table
-- Adjacency list for the social graph
CREATE TABLE IF NOT EXISTS user_follows (
    follower_id UUID NOT NULL,
    following_id UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    -- Primary key is the combination
    PRIMARY KEY (follower_id, following_id),

    -- Foreign keys
    CONSTRAINT fk_follower FOREIGN KEY (follower_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_following FOREIGN KEY (following_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE,

    -- Prevent self-following
    CONSTRAINT no_self_follow CHECK (follower_id != following_id)
);

-- Room Participants Table
-- Tracks who is in which room
CREATE TABLE IF NOT EXISTS room_participants (
    room_id UUID NOT NULL,
    user_id UUID NOT NULL,
    joined_at TIMESTAMPTZ DEFAULT NOW(),

    -- Primary key
    PRIMARY KEY (room_id, user_id),

    -- Foreign keys
    CONSTRAINT fk_room FOREIGN KEY (room_id)
        REFERENCES watch_sessions(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_participant FOREIGN KEY (user_id)
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

-- Index for votes by session
CREATE INDEX IF NOT EXISTS idx_session_votes_session
    ON session_votes(session_id);

-- Index for votes by user
CREATE INDEX IF NOT EXISTS idx_session_votes_user
    ON session_votes(user_id);

-- Index for votes by media
CREATE INDEX IF NOT EXISTS idx_session_votes_media
    ON session_votes(media_id);

-- Composite index for user votes in a session
CREATE INDEX IF NOT EXISTS idx_session_votes_session_user
    ON session_votes(session_id, user_id);

-- Index for profile username lookups
CREATE INDEX IF NOT EXISTS idx_profiles_username
    ON profiles(username);

-- Index for follower lookups
CREATE INDEX IF NOT EXISTS idx_user_follows_follower
    ON user_follows(follower_id);

-- Index for following lookups
CREATE INDEX IF NOT EXISTS idx_user_follows_following
    ON user_follows(following_id);

-- Index for room participant lookups by user
CREATE INDEX IF NOT EXISTS idx_room_participants_user
    ON room_participants(user_id);

-- Index for room participant lookups by room
CREATE INDEX IF NOT EXISTS idx_room_participants_room
    ON room_participants(room_id);

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

-- Trigger for profiles
DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;
CREATE TRIGGER update_profiles_updated_at
    BEFORE UPDATE ON profiles
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
ALTER TABLE session_votes ENABLE ROW LEVEL SECURITY;
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_follows ENABLE ROW LEVEL SECURITY;
ALTER TABLE room_participants ENABLE ROW LEVEL SECURITY;

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

-- Session Votes Policies
-- Users can read votes in sessions they created or participated in
CREATE POLICY "Users can read session votes"
    ON session_votes
    FOR SELECT
    TO authenticated
    USING (
        auth.uid() = user_id OR
        EXISTS (
            SELECT 1 FROM watch_sessions
            WHERE watch_sessions.id = session_votes.session_id
            AND watch_sessions.creator_id = auth.uid()
        )
    );

-- Users can insert their own votes
CREATE POLICY "Users can insert own votes"
    ON session_votes
    FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = user_id);

-- Users can update their own votes
CREATE POLICY "Users can update own votes"
    ON session_votes
    FOR UPDATE
    TO authenticated
    USING (auth.uid() = user_id)
    WITH CHECK (auth.uid() = user_id);

-- Users can delete their own votes
CREATE POLICY "Users can delete own votes"
    ON session_votes
    FOR DELETE
    TO authenticated
    USING (auth.uid() = user_id);

-- Profiles Policies
-- Users can read all profiles
CREATE POLICY "Users can read all profiles"
    ON profiles
    FOR SELECT
    TO authenticated
    USING (true);

-- Users can insert their own profile
CREATE POLICY "Users can insert own profile"
    ON profiles
    FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = user_id);

-- Users can update their own profile
CREATE POLICY "Users can update own profile"
    ON profiles
    FOR UPDATE
    TO authenticated
    USING (auth.uid() = user_id)
    WITH CHECK (auth.uid() = user_id);

-- User Follows Policies
-- Users can read all follows
CREATE POLICY "Users can read all follows"
    ON user_follows
    FOR SELECT
    TO authenticated
    USING (true);

-- Users can follow others
CREATE POLICY "Users can follow others"
    ON user_follows
    FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = follower_id);

-- Users can unfollow
CREATE POLICY "Users can unfollow"
    ON user_follows
    FOR DELETE
    TO authenticated
    USING (auth.uid() = follower_id);

-- Room Participants Policies
-- Users can read room participants if they are in the room
CREATE POLICY "Users can read room participants"
    ON room_participants
    FOR SELECT
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM room_participants rp
            WHERE rp.room_id = room_participants.room_id
            AND rp.user_id = auth.uid()
        )
    );

-- Room creators can add participants
CREATE POLICY "Room creators can add participants"
    ON room_participants
    FOR INSERT
    TO authenticated
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM watch_sessions
            WHERE watch_sessions.id = room_participants.room_id
            AND watch_sessions.creator_id = auth.uid()
        )
    );

-- Users can leave rooms
CREATE POLICY "Users can leave rooms"
    ON room_participants
    FOR DELETE
    TO authenticated
    USING (auth.uid() = user_id);

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

COMMENT ON TABLE session_votes IS 'Stores user votes for media items within watch sessions';
COMMENT ON COLUMN session_votes.vote IS 'User vote: yes, no, or maybe';
COMMENT ON COLUMN session_votes.session_id IS 'Watch session this vote belongs to';
COMMENT ON COLUMN session_votes.user_id IS 'User who cast this vote (references auth.users)';
COMMENT ON COLUMN session_votes.media_id IS 'Media item being voted on';

COMMENT ON TABLE profiles IS 'User profile information and privacy settings';
COMMENT ON COLUMN profiles.username IS 'Unique username for the user';
COMMENT ON COLUMN profiles.invite_preference IS 'Privacy setting for room invitations';

COMMENT ON TABLE user_follows IS 'Social graph adjacency list for follower relationships';
COMMENT ON COLUMN user_follows.follower_id IS 'User who is following';
COMMENT ON COLUMN user_follows.following_id IS 'User who is being followed';

COMMENT ON TABLE room_participants IS 'Tracks which users are in which rooms';
COMMENT ON COLUMN room_participants.room_id IS 'Room (watch session) the user is in';
COMMENT ON COLUMN room_participants.user_id IS 'User participating in the room';
