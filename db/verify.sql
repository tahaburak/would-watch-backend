-- Would Watch Backend - Schema Verification Script
-- Run this after applying schema.sql to verify everything is set up correctly

-- ============================================================================
-- CHECK TYPES
-- ============================================================================

DO $$
BEGIN
    -- Check media_type enum
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'media_type') THEN
        RAISE NOTICE '✓ Type media_type exists';
    ELSE
        RAISE WARNING '✗ Type media_type is missing';
    END IF;

    -- Check session_status enum
    IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'session_status') THEN
        RAISE NOTICE '✓ Type session_status exists';
    ELSE
        RAISE WARNING '✗ Type session_status is missing';
    END IF;
END $$;

-- ============================================================================
-- CHECK TABLES
-- ============================================================================

DO $$
BEGIN
    -- Check media_items table
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'media_items') THEN
        RAISE NOTICE '✓ Table media_items exists';
    ELSE
        RAISE WARNING '✗ Table media_items is missing';
    END IF;

    -- Check watch_sessions table
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'watch_sessions') THEN
        RAISE NOTICE '✓ Table watch_sessions exists';
    ELSE
        RAISE WARNING '✗ Table watch_sessions is missing';
    END IF;
END $$;

-- ============================================================================
-- CHECK CONSTRAINTS
-- ============================================================================

DO $$
BEGIN
    -- Check unique constraint on media_items
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'media_items'
        AND constraint_name = 'unique_tmdb_media'
    ) THEN
        RAISE NOTICE '✓ Unique constraint unique_tmdb_media exists';
    ELSE
        RAISE WARNING '✗ Unique constraint unique_tmdb_media is missing';
    END IF;

    -- Check foreign key on watch_sessions
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE table_name = 'watch_sessions'
        AND constraint_name = 'fk_creator'
    ) THEN
        RAISE NOTICE '✓ Foreign key fk_creator exists';
    ELSE
        RAISE WARNING '✗ Foreign key fk_creator is missing';
    END IF;
END $$;

-- ============================================================================
-- CHECK INDEXES
-- ============================================================================

DO $$
DECLARE
    index_names TEXT[] := ARRAY[
        'idx_media_items_tmdb_id',
        'idx_media_items_type',
        'idx_media_items_metadata',
        'idx_watch_sessions_status',
        'idx_watch_sessions_creator',
        'idx_watch_sessions_creator_status',
        'idx_watch_sessions_created'
    ];
    idx TEXT;
BEGIN
    FOREACH idx IN ARRAY index_names
    LOOP
        IF EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = idx) THEN
            RAISE NOTICE '✓ Index % exists', idx;
        ELSE
            RAISE WARNING '✗ Index % is missing', idx;
        END IF;
    END LOOP;
END $$;

-- ============================================================================
-- CHECK TRIGGERS
-- ============================================================================

DO $$
BEGIN
    -- Check media_items trigger
    IF EXISTS (
        SELECT 1 FROM information_schema.triggers
        WHERE trigger_name = 'update_media_items_updated_at'
    ) THEN
        RAISE NOTICE '✓ Trigger update_media_items_updated_at exists';
    ELSE
        RAISE WARNING '✗ Trigger update_media_items_updated_at is missing';
    END IF;

    -- Check watch_sessions update trigger
    IF EXISTS (
        SELECT 1 FROM information_schema.triggers
        WHERE trigger_name = 'update_watch_sessions_updated_at'
    ) THEN
        RAISE NOTICE '✓ Trigger update_watch_sessions_updated_at exists';
    ELSE
        RAISE WARNING '✗ Trigger update_watch_sessions_updated_at is missing';
    END IF;

    -- Check watch_sessions completed trigger
    IF EXISTS (
        SELECT 1 FROM information_schema.triggers
        WHERE trigger_name = 'set_watch_session_completed_at'
    ) THEN
        RAISE NOTICE '✓ Trigger set_watch_session_completed_at exists';
    ELSE
        RAISE WARNING '✗ Trigger set_watch_session_completed_at is missing';
    END IF;
END $$;

-- ============================================================================
-- CHECK FUNCTIONS
-- ============================================================================

DO $$
BEGIN
    -- Check update_updated_at_column function
    IF EXISTS (
        SELECT 1 FROM pg_proc
        WHERE proname = 'update_updated_at_column'
    ) THEN
        RAISE NOTICE '✓ Function update_updated_at_column exists';
    ELSE
        RAISE WARNING '✗ Function update_updated_at_column is missing';
    END IF;

    -- Check set_completed_at function
    IF EXISTS (
        SELECT 1 FROM pg_proc
        WHERE proname = 'set_completed_at'
    ) THEN
        RAISE NOTICE '✓ Function set_completed_at exists';
    ELSE
        RAISE WARNING '✗ Function set_completed_at is missing';
    END IF;
END $$;

-- ============================================================================
-- CHECK RLS
-- ============================================================================

DO $$
BEGIN
    -- Check RLS is enabled on media_items
    IF EXISTS (
        SELECT 1 FROM pg_tables
        WHERE tablename = 'media_items'
        AND rowsecurity = true
    ) THEN
        RAISE NOTICE '✓ RLS enabled on media_items';
    ELSE
        RAISE WARNING '✗ RLS not enabled on media_items';
    END IF;

    -- Check RLS is enabled on watch_sessions
    IF EXISTS (
        SELECT 1 FROM pg_tables
        WHERE tablename = 'watch_sessions'
        AND rowsecurity = true
    ) THEN
        RAISE NOTICE '✓ RLS enabled on watch_sessions';
    ELSE
        RAISE WARNING '✗ RLS not enabled on watch_sessions';
    END IF;
END $$;

-- ============================================================================
-- SHOW TABLE STRUCTURE
-- ============================================================================

\echo ''
\echo '========================================='
\echo 'Table: media_items'
\echo '========================================='
\d media_items

\echo ''
\echo '========================================='
\echo 'Table: watch_sessions'
\echo '========================================='
\d watch_sessions

-- ============================================================================
-- SHOW POLICIES
-- ============================================================================

\echo ''
\echo '========================================='
\echo 'RLS Policies'
\echo '========================================='
SELECT schemaname, tablename, policyname, permissive, roles, cmd
FROM pg_policies
WHERE tablename IN ('media_items', 'watch_sessions')
ORDER BY tablename, policyname;
