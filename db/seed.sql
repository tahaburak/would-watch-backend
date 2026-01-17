-- Would Watch Backend - Seed Data
-- Sample data for testing and development
-- NOTE: This assumes you have a test user in auth.users table

-- ============================================================================
-- SAMPLE MEDIA ITEMS
-- ============================================================================

-- Insert some popular movies
INSERT INTO media_items (tmdb_id, media_type, title, metadata) VALUES
(550, 'movie', 'Fight Club', '{
    "poster_path": "/pB8BM7pdSp6B6Ih7QZ4DrQ3PmJK.jpg",
    "backdrop_path": "/hZkgoQYus5vegHoetLkCJzb17zJ.jpg",
    "overview": "A ticking-time-bomb insomniac and a slippery soap salesman channel primal male aggression into a shocking new form of therapy.",
    "release_date": "1999-10-15",
    "vote_average": 8.4,
    "vote_count": 26000,
    "runtime": 139,
    "genres": [{"id": 18, "name": "Drama"}]
}'::jsonb),

(13, 'movie', 'Forrest Gump', '{
    "poster_path": "/arw2vcBveWOVZr6pxd9XTd1TdQa.jpg",
    "backdrop_path": "/7c9UVPPiTPltouxRVY6N9uAIm8P.jpg",
    "overview": "A man with a low IQ has accomplished great things in his life and been present during significant historic events.",
    "release_date": "1994-07-06",
    "vote_average": 8.5,
    "vote_count": 24000,
    "runtime": 142,
    "genres": [{"id": 18, "name": "Drama"}, {"id": 10749, "name": "Romance"}]
}'::jsonb),

(278, 'movie', 'The Shawshank Redemption', '{
    "poster_path": "/q6y0Go1tsGEsmtFryDOJo3dEmqu.jpg",
    "backdrop_path": "/kXfqcdQKsToO0OUXHcrrNCHDBzO.jpg",
    "overview": "Framed for murder, upstanding banker Andy Dufresne begins a new life at Shawshank prison.",
    "release_date": "1994-09-23",
    "vote_average": 8.7,
    "vote_count": 23000,
    "runtime": 142,
    "genres": [{"id": 18, "name": "Drama"}, {"id": 80, "name": "Crime"}]
}'::jsonb),

(603, 'movie', 'The Matrix', '{
    "poster_path": "/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg",
    "backdrop_path": "/fNG7i7RqMErkcqhohV2a6cV1Ehy.jpg",
    "overview": "A computer hacker learns about the true nature of reality and his role in the war against its controllers.",
    "release_date": "1999-03-31",
    "vote_average": 8.2,
    "vote_count": 22000,
    "runtime": 136,
    "genres": [{"id": 28, "name": "Action"}, {"id": 878, "name": "Science Fiction"}]
}'::jsonb),

(680, 'movie', 'Pulp Fiction', '{
    "poster_path": "/d5iIlFn5s0ImszYzBPb8JPIfbXD.jpg",
    "backdrop_path": "/4cDFJr4HnXN5AdPw4AKrmLlMWdO.jpg",
    "overview": "The lives of two mob hitmen, a boxer, a gangster and his wife intertwine in four tales of violence and redemption.",
    "release_date": "1994-09-10",
    "vote_average": 8.5,
    "vote_count": 25000,
    "runtime": 154,
    "genres": [{"id": 53, "name": "Thriller"}, {"id": 80, "name": "Crime"}]
}'::jsonb),

(155, 'movie', 'The Dark Knight', '{
    "poster_path": "/qJ2tW6WMUDux911r6m7haRef0WH.jpg",
    "backdrop_path": "/hkBaDkMWbLaf8B1lsWsKX7Ew3Xq.jpg",
    "overview": "Batman raises the stakes in his war on crime with the help of Lt. Jim Gordon and District Attorney Harvey Dent.",
    "release_date": "2008-07-18",
    "vote_average": 8.5,
    "vote_count": 28000,
    "runtime": 152,
    "genres": [{"id": 18, "name": "Drama"}, {"id": 28, "name": "Action"}, {"id": 80, "name": "Crime"}]
}'::jsonb),

(238, 'movie', 'The Godfather', '{
    "poster_path": "/3bhkrj58Vtu7enYsRolD1fZdja1.jpg",
    "backdrop_path": "/tmU7GeKVybMWFButWEGl2M4GeiP.jpg",
    "overview": "The aging patriarch of an organized crime dynasty transfers control of his clandestine empire to his reluctant son.",
    "release_date": "1972-03-14",
    "vote_average": 8.7,
    "vote_count": 17000,
    "runtime": 175,
    "genres": [{"id": 18, "name": "Drama"}, {"id": 80, "name": "Crime"}]
}'::jsonb)

ON CONFLICT (tmdb_id, media_type) DO NOTHING;

-- Insert some TV shows
INSERT INTO media_items (tmdb_id, media_type, title, metadata) VALUES
(1396, 'tv', 'Breaking Bad', '{
    "poster_path": "/ggFHVNu6YYI5L9pCfOacjizRGt.jpg",
    "backdrop_path": "/9faGSFi5jam6pDWGNd0p8JcJgXQ.jpg",
    "overview": "A high school chemistry teacher diagnosed with terminal lung cancer teams up with a former student to manufacture meth.",
    "first_air_date": "2008-01-20",
    "vote_average": 8.9,
    "vote_count": 11000,
    "genres": [{"id": 18, "name": "Drama"}]
}'::jsonb),

(1399, 'tv', 'Game of Thrones', '{
    "poster_path": "/u3bZgnGQ9T01sWNhyveQz0wH0Hl.jpg",
    "backdrop_path": "/suopoADq0k8YZr4dQXcU6pToj6s.jpg",
    "overview": "Nine noble families fight for control over the lands of Westeros while an ancient enemy returns.",
    "first_air_date": "2011-04-17",
    "vote_average": 8.4,
    "vote_count": 20000,
    "genres": [{"id": 18, "name": "Drama"}, {"id": 10765, "name": "Sci-Fi & Fantasy"}]
}'::jsonb),

(94605, 'tv', 'Arcane', '{
    "poster_path": "/fqldf2t8ztc9aiwn3k6mlX3tvRT.jpg",
    "backdrop_path": "/rkB4LyZHo1NHXFEDHl9vSD9r1lI.jpg",
    "overview": "Amid the stark discord of twin cities Piltover and Zaun, two sisters fight on rival sides of a war between magic technologies and clashing convictions.",
    "first_air_date": "2021-11-06",
    "vote_average": 8.8,
    "vote_count": 3000,
    "genres": [{"id": 16, "name": "Animation"}, {"id": 10765, "name": "Sci-Fi & Fantasy"}]
}'::jsonb)

ON CONFLICT (tmdb_id, media_type) DO NOTHING;

-- ============================================================================
-- SAMPLE WATCH SESSIONS
-- ============================================================================

-- NOTE: Replace 'YOUR_USER_ID_HERE' with an actual user UUID from auth.users
-- You can get this by running: SELECT id FROM auth.users LIMIT 1;

-- To insert sample sessions, uncomment and modify:
/*
INSERT INTO watch_sessions (creator_id, status) VALUES
('YOUR_USER_ID_HERE', 'active'),
('YOUR_USER_ID_HERE', 'completed');
*/

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Show inserted media items
SELECT
    id,
    tmdb_id,
    media_type,
    title,
    metadata->>'vote_average' as rating,
    created_at
FROM media_items
ORDER BY media_type, title;

-- Show total counts
SELECT
    media_type,
    COUNT(*) as count
FROM media_items
GROUP BY media_type;

DO $$
DECLARE
    movie_count INTEGER;
    tv_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO movie_count FROM media_items WHERE media_type = 'movie';
    SELECT COUNT(*) INTO tv_count FROM media_items WHERE media_type = 'tv';

    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Seed data inserted successfully!';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Movies: %', movie_count;
    RAISE NOTICE 'TV Shows: %', tv_count;
    RAISE NOTICE 'Total: %', movie_count + tv_count;
    RAISE NOTICE '';
    RAISE NOTICE 'To create sample sessions, update the';
    RAISE NOTICE 'commented section with your user ID.';
    RAISE NOTICE '========================================';
END $$;
