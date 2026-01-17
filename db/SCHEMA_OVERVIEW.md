# Database Schema Overview

Quick reference for the Would Watch database schema.

## Table Relationships

```
┌─────────────────┐
│   auth.users    │ (Supabase managed)
│   (External)    │
└────────┬────────┘
         │
         │ creator_id (FK)
         │
         ▼
┌─────────────────────────┐
│   watch_sessions        │
├─────────────────────────┤
│ id (PK)                 │
│ creator_id (FK)         │──┐
│ status (ENUM)           │  │
│ created_at              │  │
│ updated_at              │  │
│ completed_at (nullable) │  │
└─────────────────────────┘  │
                             │
                             │ (Future: session_votes)
                             │
                             ▼
                    ┌─────────────────┐
                    │  media_items    │
                    ├─────────────────┤
                    │ id (PK)         │
                    │ tmdb_id         │
                    │ media_type      │
                    │ title           │
                    │ metadata (JSONB)│
                    │ created_at      │
                    │ updated_at      │
                    └─────────────────┘
```

## Table Stats

| Table          | Columns | Indexes | Triggers | Policies |
|----------------|---------|---------|----------|----------|
| media_items    | 6       | 3       | 1        | 2        |
| watch_sessions | 6       | 4       | 2        | 4        |

## Enum Types

### media_type
- `movie` - Feature films
- `tv` - TV shows/series

### session_status
- `active` - Session accepting votes
- `completed` - Session finished

## Key Features

### 🔒 Security
- Row Level Security (RLS) enabled on all tables
- Users can only access their own sessions
- JWT-based authentication via Supabase Auth

### ⚡ Performance
- 7 strategically placed indexes
- GIN index on JSONB metadata for fast queries
- Composite indexes for common query patterns

### 🔄 Automation
- Auto-updating timestamps via triggers
- Automatic `completed_at` setting
- UUID generation for primary keys

### 🎯 Data Integrity
- Unique constraint on (tmdb_id, media_type)
- Foreign key to auth.users with CASCADE delete
- NOT NULL constraints on critical fields

## Common Operations

### Query Active Sessions
```sql
SELECT * FROM watch_sessions
WHERE creator_id = auth.uid()
  AND status = 'active';
```

### Find Movies by Rating
```sql
SELECT title, metadata->>'vote_average' as rating
FROM media_items
WHERE media_type = 'movie'
  AND (metadata->>'vote_average')::float >= 8.0
ORDER BY (metadata->>'vote_average')::float DESC;
```

### Upsert Media Item
```sql
INSERT INTO media_items (tmdb_id, media_type, title, metadata)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tmdb_id, media_type)
DO UPDATE SET
    title = EXCLUDED.title,
    metadata = EXCLUDED.metadata;
```

## Migration Files

| File          | Purpose                                    |
|---------------|--------------------------------------------|
| schema.sql    | Create all tables, indexes, policies      |
| rollback.sql  | Remove all schema objects                 |
| verify.sql    | Verify schema was applied correctly       |
| seed.sql      | Insert sample data for testing            |

## Quick Start

1. **Apply Schema**
   ```bash
   psql $DATABASE_URL -f backend/db/schema.sql
   ```

2. **Verify**
   ```bash
   psql $DATABASE_URL -f backend/db/verify.sql
   ```

3. **Seed Data** (optional)
   ```bash
   psql $DATABASE_URL -f backend/db/seed.sql
   ```

4. **Rollback** (if needed)
   ```bash
   psql $DATABASE_URL -f backend/db/rollback.sql
   ```

## Future Enhancements

Potential tables for future sprints:

- **session_participants** - Track users in a session
- **session_votes** - Individual votes for media items
- **user_watch_history** - Track viewing history
- **user_preferences** - Store genre preferences
- **notifications** - Session invites and updates
- **comments** - Comments on media items

## Indexes Explained

| Index Name                          | Purpose                           |
|-------------------------------------|-----------------------------------|
| idx_media_items_tmdb_id             | Fast lookups by TMDB ID           |
| idx_media_items_type                | Filter by movie/tv                |
| idx_media_items_metadata (GIN)      | Query JSONB fields efficiently    |
| idx_watch_sessions_status           | Filter by active/completed        |
| idx_watch_sessions_creator          | Get user's sessions               |
| idx_watch_sessions_creator_status   | Combined creator + status queries |
| idx_watch_sessions_created          | Sort by recent sessions           |
