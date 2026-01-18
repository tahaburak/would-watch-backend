-- Would Watch Backend - Database Schema
-- This schema should be run against your Supabase PostgreSQL database

-- ============================================================================
-- TYPES
-- ============================================================================

-- Create enum for media types
CREATE TYPE IF NOT EXISTS media_type AS ENUM ('movie', 'tv');

-- Create enum for session status
CREATE TYPE IF NOT EXISTS session_status AS ENUM ('active', 'completed');

-- Create enum for vote types
CREATE TYPE IF NOT EXISTS vote_type AS ENUM ('yes', 'no', 'maybe');

-- Create enum for invite preference
CREATE TYPE IF NOT EXISTS invite_preference AS ENUM ('everyone', 'following', 'none');

-- Create enum for room participant roles
CREATE TYPE IF NOT EXISTS participant_role AS ENUM ('owner', 'admin', 'viewer');

-- Create enum for participant status
CREATE TYPE IF NOT EXISTS participant_status AS ENUM ('invited', 'joined', 'declined');

-- ============================================================================
-- TABLES
-- ============================================================================

-- Profiles Table (Public Profile Data & Settings)
CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    username TEXT UNIQUE,
    avatar_url TEXT,
    invite_preference invite_preference NOT NULL DEFAULT 'following',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Follows Table (Social Graph)
CREATE TABLE IF NOT EXISTS user_follows (
    follower_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (follower_id, following_id),
    CONSTRAINT no_self_follow CHECK (follower_id != following_id)
);

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

-- Watch Sessions Table (Rooms)
-- Stores group watch sessions for collaborative movie selection
CREATE TABLE IF NOT EXISTS watch_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    status session_status NOT NULL DEFAULT 'active',
    name TEXT, -- Optional room name
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Room Participants Table
CREATE TABLE IF NOT EXISTS room_participants (
    room_id UUID NOT NULL REFERENCES watch_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    role participant_role NOT NULL DEFAULT 'viewer',
    status participant_status NOT NULL DEFAULT 'invited',
    joined_at TIMESTAMPTZ,

    PRIMARY KEY (room_id, user_id)
);

-- Session Votes Table
-- Stores user votes for media items within watch sessions
CREATE TABLE IF NOT EXISTS session_votes (
    session_id UUID NOT NULL REFERENCES watch_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    media_id UUID NOT NULL REFERENCES media_items(id) ON DELETE CASCADE,
    vote vote_type NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Unique constraint: one vote per user per media per session
    CONSTRAINT unique_user_media_session UNIQUE (session_id, user_id, media_id)
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

-- Trigger for session_votes
DROP TRIGGER IF EXISTS update_session_votes_updated_at ON session_votes;
CREATE TRIGGER update_session_votes_updated_at
    BEFORE UPDATE ON session_votes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function and Trigger to create profile on Signup
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.profiles (id, username, avatar_url)
  VALUES (
    NEW.id,
    NEW.raw_user_meta_data->>'username', -- Can be null initially
    NEW.raw_user_meta_data->>'avatar_url'
  );
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Trigger to auto-create profile when user signs up
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- ============================================================================
-- ROW LEVEL SECURITY POLICIES
-- ============================================================================

-- Enable RLS on all tables
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_follows ENABLE ROW LEVEL SECURITY;
ALTER TABLE room_participants ENABLE ROW LEVEL SECURITY;
ALTER TABLE watch_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE session_votes ENABLE ROW LEVEL SECURITY;
ALTER TABLE media_items ENABLE ROW LEVEL SECURITY;

-- Profiles Policies
DROP POLICY IF EXISTS "Users can read all profiles" ON profiles;
CREATE POLICY "Users can read all profiles"
    ON profiles
    FOR SELECT
    TO authenticated
    USING (true);

DROP POLICY IF EXISTS "Users can insert own profile" ON profiles;
CREATE POLICY "Users can insert own profile"
    ON profiles
    FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = id);

DROP POLICY IF EXISTS "Users can update own profile" ON profiles;
CREATE POLICY "Users can update own profile"
    ON profiles
    FOR UPDATE
    TO authenticated
    USING (auth.uid() = id)
    WITH CHECK (auth.uid() = id);

-- User Follows Policies
DROP POLICY IF EXISTS "Users can read all follows" ON user_follows;
CREATE POLICY "Users can read all follows"
    ON user_follows
    FOR SELECT
    TO authenticated
    USING (true);

DROP POLICY IF EXISTS "Users can follow others" ON user_follows;
CREATE POLICY "Users can follow others"
    ON user_follows
    FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = follower_id);

DROP POLICY IF EXISTS "Users can unfollow" ON user_follows;
CREATE POLICY "Users can unfollow"
    ON user_follows
    FOR DELETE
    TO authenticated
    USING (auth.uid() = follower_id);

-- Watch Sessions Policies
DROP POLICY IF EXISTS "Users can read public sessions" ON watch_sessions;
CREATE POLICY "Users can read public sessions"
    ON watch_sessions
    FOR SELECT
    TO authenticated
    USING (
        is_public = true OR
        creator_id = auth.uid() OR
        EXISTS (
            SELECT 1 FROM room_participants
            WHERE room_participants.room_id = watch_sessions.id
            AND room_participants.user_id = auth.uid()
        )
    );

DROP POLICY IF EXISTS "Users can create sessions" ON watch_sessions;
CREATE POLICY "Users can create sessions"
    ON watch_sessions
    FOR INSERT
    TO authenticated
    WITH CHECK (auth.uid() = creator_id);

DROP POLICY IF EXISTS "Creators can update sessions" ON watch_sessions;
CREATE POLICY "Creators can update sessions"
    ON watch_sessions
    FOR UPDATE
    TO authenticated
    USING (auth.uid() = creator_id)
    WITH CHECK (auth.uid() = creator_id);

DROP POLICY IF EXISTS "Creators can delete sessions" ON watch_sessions;
CREATE POLICY "Creators can delete sessions"
    ON watch_sessions
    FOR DELETE
    TO authenticated
    USING (auth.uid() = creator_id);

-- Room Participants Policies
DROP POLICY IF EXISTS "Users can read room participants" ON room_participants;
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

DROP POLICY IF EXISTS "Room creators can add participants" ON room_participants;
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

DROP POLICY IF EXISTS "Users can leave rooms" ON room_participants;
CREATE POLICY "Users can leave rooms"
    ON room_participants
    FOR DELETE
    TO authenticated
    USING (auth.uid() = user_id);

-- Session Votes Policies
DROP POLICY IF EXISTS "Users can read session votes" ON session_votes;
CREATE POLICY "Users can read session votes"
    ON session_votes
    FOR SELECT
    TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM room_participants
            WHERE room_participants.room_id = session_votes.session_id
            AND room_participants.user_id = auth.uid()
        )
    );

DROP POLICY IF EXISTS "Users can insert session votes" ON session_votes;
CREATE POLICY "Users can insert session votes"
    ON session_votes
    FOR INSERT
    TO authenticated
    WITH CHECK (
        auth.uid() = user_id AND
        EXISTS (
            SELECT 1 FROM room_participants
            WHERE room_participants.room_id = session_votes.session_id
            AND room_participants.user_id = auth.uid()
        )
    );

DROP POLICY IF EXISTS "Users can update own votes" ON session_votes;
CREATE POLICY "Users can update own votes"
    ON session_votes
    FOR UPDATE
    TO authenticated
    USING (auth.uid() = user_id)
    WITH CHECK (auth.uid() = user_id);

-- Media Items Policies
DROP POLICY IF EXISTS "Users can read all media" ON media_items;
CREATE POLICY "Users can read all media"
    ON media_items
    FOR SELECT
    TO authenticated
    USING (true);

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
COMMENT ON COLUMN session_votes.user_id IS 'User who cast this vote (references profiles)';
COMMENT ON COLUMN session_votes.media_id IS 'Media item being voted on';

COMMENT ON TABLE profiles IS 'User profile information and privacy settings';
COMMENT ON COLUMN profiles.id IS 'User ID (references auth.users)';
COMMENT ON COLUMN profiles.username IS 'Unique username for the user';
COMMENT ON COLUMN profiles.invite_preference IS 'Privacy setting for room invitations';

COMMENT ON TABLE user_follows IS 'Social graph adjacency list for follower relationships';
COMMENT ON COLUMN user_follows.follower_id IS 'User who is following (references profiles)';
COMMENT ON COLUMN user_follows.following_id IS 'User who is being followed (references profiles)';

COMMENT ON TABLE room_participants IS 'Tracks which users are in which rooms';
COMMENT ON COLUMN room_participants.room_id IS 'Room (watch session) the user is in';
COMMENT ON COLUMN room_participants.user_id IS 'User participating in the room (references profiles)';
COMMENT ON COLUMN room_participants.role IS 'Participant role: owner, admin, or viewer';
COMMENT ON COLUMN room_participants.status IS 'Participant status: invited, joined, or declined';
