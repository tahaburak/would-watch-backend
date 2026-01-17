-- Would Watch Backend - Rollback Script
-- Use this to remove all tables, types, functions, and policies created by schema.sql
-- WARNING: This will permanently delete all data in these tables!

-- ============================================================================
-- DROP POLICIES
-- ============================================================================

-- Drop media_items policies
DROP POLICY IF EXISTS "Allow authenticated users to read media items" ON media_items;
DROP POLICY IF EXISTS "Allow authenticated users to insert media items" ON media_items;

-- Drop watch_sessions policies
DROP POLICY IF EXISTS "Users can read own sessions" ON watch_sessions;
DROP POLICY IF EXISTS "Users can create sessions" ON watch_sessions;
DROP POLICY IF EXISTS "Users can update own sessions" ON watch_sessions;
DROP POLICY IF EXISTS "Users can delete own sessions" ON watch_sessions;

-- ============================================================================
-- DROP TRIGGERS
-- ============================================================================

DROP TRIGGER IF EXISTS update_media_items_updated_at ON media_items;
DROP TRIGGER IF EXISTS update_watch_sessions_updated_at ON watch_sessions;
DROP TRIGGER IF EXISTS set_watch_session_completed_at ON watch_sessions;

-- ============================================================================
-- DROP FUNCTIONS
-- ============================================================================

DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS set_completed_at();

-- ============================================================================
-- DROP TABLES
-- ============================================================================

-- Drop tables (CASCADE will also drop dependent objects)
DROP TABLE IF EXISTS watch_sessions CASCADE;
DROP TABLE IF EXISTS media_items CASCADE;

-- ============================================================================
-- DROP TYPES
-- ============================================================================

DROP TYPE IF EXISTS session_status;
DROP TYPE IF EXISTS media_type;

-- ============================================================================
-- CONFIRMATION
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Rollback completed successfully. All schema objects have been removed.';
END $$;
