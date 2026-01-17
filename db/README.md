# Database Schema

This directory contains the SQL schema for the Would Watch application backend.

## Overview

The database schema consists of two main tables:
1. **media_items** - Stores movie and TV show information from TMDB
2. **watch_sessions** - Manages group watch sessions for collaborative movie selection

## Running the Migration

### Using Supabase Dashboard

1. Log in to your [Supabase Dashboard](https://app.supabase.com)
2. Navigate to your project
3. Go to **SQL Editor** in the left sidebar
4. Click **New Query**
5. Copy and paste the contents of `schema.sql`
6. Click **Run** or press `Ctrl/Cmd + Enter`

### Using Supabase CLI

```bash
# Install Supabase CLI if not already installed
npm install -g supabase

# Login to Supabase
supabase login

# Link to your project
supabase link --project-ref your-project-ref

# Run the migration
supabase db push

# Or apply the schema directly
psql $DATABASE_URL -f db/schema.sql
```

## Schema Details

### Tables

#### media_items

Stores movie and TV show metadata fetched from TMDB API.

| Column      | Type      | Description                                    |
|-------------|-----------|------------------------------------------------|
| id          | UUID      | Primary key, auto-generated                    |
| tmdb_id     | INTEGER   | The Movie Database ID                          |
| media_type  | ENUM      | 'movie' or 'tv'                                |
| title       | TEXT      | Movie or show title                            |
| metadata    | JSONB     | Flexible JSON storage for poster, overview, etc|
| created_at  | TIMESTAMPTZ | When the record was created                  |
| updated_at  | TIMESTAMPTZ | Last update timestamp (auto-updated)         |

**Constraints:**
- Unique constraint on `(tmdb_id, media_type)` to prevent duplicates

**Indexes:**
- `idx_media_items_tmdb_id` - Fast TMDB ID lookups
- `idx_media_items_type` - Filter by media type
- `idx_media_items_metadata` - GIN index for JSONB queries

**Example metadata structure:**
```json
{
  "poster_path": "/path/to/poster.jpg",
  "backdrop_path": "/path/to/backdrop.jpg",
  "overview": "Movie description...",
  "release_date": "2024-01-15",
  "vote_average": 8.5,
  "vote_count": 1000,
  "genres": [{"id": 28, "name": "Action"}],
  "runtime": 120
}
```

#### watch_sessions

Manages group watch sessions where friends vote on movies to watch together.

| Column       | Type        | Description                                    |
|--------------|-------------|------------------------------------------------|
| id           | UUID        | Primary key, auto-generated                    |
| creator_id   | UUID        | Foreign key to auth.users                      |
| status       | ENUM        | 'active' or 'completed'                        |
| created_at   | TIMESTAMPTZ | When the session was created                   |
| updated_at   | TIMESTAMPTZ | Last update timestamp (auto-updated)           |
| completed_at | TIMESTAMPTZ | When session was marked completed (nullable)   |

**Constraints:**
- Foreign key to `auth.users(id)` with CASCADE delete

**Indexes:**
- `idx_watch_sessions_status` - Filter by session status
- `idx_watch_sessions_creator` - Find user's sessions
- `idx_watch_sessions_creator_status` - Composite index for filtered queries
- `idx_watch_sessions_created` - Sort by creation date

### Enums

#### media_type
- `movie` - Feature films
- `tv` - TV shows/series

#### session_status
- `active` - Session is currently accepting votes
- `completed` - Session has ended and a movie was selected

### Triggers

1. **update_updated_at_column**
   - Automatically updates `updated_at` timestamp on any UPDATE
   - Applied to: `media_items`, `watch_sessions`

2. **set_completed_at**
   - Automatically sets `completed_at` when status changes to 'completed'
   - Applied to: `watch_sessions`

### Row Level Security (RLS)

RLS is enabled on all tables to ensure data security.

#### media_items Policies
- **Read**: All authenticated users can read media items
- **Insert**: All authenticated users can insert media items

#### watch_sessions Policies
- **Read**: Users can only read their own sessions
- **Insert**: Users can create sessions (enforces creator_id = auth.uid())
- **Update**: Users can only update their own sessions
- **Delete**: Users can only delete their own sessions

## Common Queries

### Insert a movie from TMDB

```sql
INSERT INTO media_items (tmdb_id, media_type, title, metadata)
VALUES (
    550,
    'movie',
    'Fight Club',
    '{
        "poster_path": "/path.jpg",
        "overview": "Description...",
        "release_date": "1999-10-15",
        "vote_average": 8.4
    }'::jsonb
)
ON CONFLICT (tmdb_id, media_type)
DO UPDATE SET
    title = EXCLUDED.title,
    metadata = EXCLUDED.metadata;
```

### Create a watch session

```sql
INSERT INTO watch_sessions (creator_id, status)
VALUES (
    'user-uuid-here',
    'active'
)
RETURNING id;
```

### Get user's active sessions

```sql
SELECT *
FROM watch_sessions
WHERE creator_id = 'user-uuid-here'
  AND status = 'active'
ORDER BY created_at DESC;
```

### Mark session as completed

```sql
UPDATE watch_sessions
SET status = 'completed'
WHERE id = 'session-uuid-here'
  AND creator_id = 'user-uuid-here';
-- completed_at will be set automatically by trigger
```

### Search media items by title

```sql
SELECT *
FROM media_items
WHERE title ILIKE '%matrix%'
  AND media_type = 'movie';
```

### Query JSONB metadata

```sql
-- Find highly-rated movies
SELECT title, metadata->>'vote_average' as rating
FROM media_items
WHERE media_type = 'movie'
  AND (metadata->>'vote_average')::float > 8.0;

-- Find movies by genre
SELECT *
FROM media_items
WHERE metadata @> '{"genres": [{"name": "Action"}]}';
```

## Future Extensions

Consider adding these tables in future sprints:

- **session_participants** - Track users participating in a session
- **session_votes** - Store individual votes for movies in a session
- **user_watch_history** - Track what users have watched
- **user_preferences** - Store user preferences for recommendations

## Maintenance

### Viewing table information

```sql
-- Show all tables
\dt

-- Describe a table
\d media_items

-- Show indexes
\di

-- Show RLS policies
\dp media_items
```

### Backup and Restore

```bash
# Backup
pg_dump $DATABASE_URL > backup.sql

# Restore
psql $DATABASE_URL < backup.sql
```

## Troubleshooting

### Permission Issues

If you get permission errors, ensure:
1. You're running queries as an authenticated user
2. RLS policies are correctly configured
3. The user has the correct role (authenticated vs anon)

### Duplicate Key Errors

If you get unique constraint violations on `(tmdb_id, media_type)`:
- Use `ON CONFLICT` clause to handle upserts
- Check if the media item already exists before inserting

### Foreign Key Violations

If you get FK violations on `creator_id`:
- Ensure the user exists in `auth.users`
- Use valid UUIDs from Supabase Auth
