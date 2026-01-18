package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/tahaburak/would-watch-backend/internal/database"
	"github.com/tahaburak/would-watch-backend/internal/middleware"
	"github.com/tahaburak/would-watch-backend/internal/testutils"
)

// TestServer wraps the test HTTP server and database
type TestServer struct {
	Server     *httptest.Server
	DB         *testutils.TestDB
	Expect     *httpexpect.Expect
	MockUserID string
}

// NewTestServer creates a new test server with all handlers configured
func NewTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Create test database
	testDB := testutils.NewTestDB(t)

	// Initialize Repositories
	roomRepo := database.NewRoomRepository(testDB.DB)
	socialRepo := database.NewSocialRepository(testDB.DB)
	sessionRepo := database.NewSessionRepository(testDB.DB)
	voteRepo := database.NewVoteRepository(testDB.DB)

	// Initialize Handlers
	roomHandler := NewRoomHandler(roomRepo, socialRepo)
	socialHandler := NewSocialHandler(socialRepo)
	sessionHandler := NewSessionHandler(sessionRepo)
	voteHandler := NewVoteHandler(voteRepo, sessionRepo)
	matchHandler := NewMatchHandler(voteRepo)

	// Create router
	mux := http.NewServeMux()

	// Mock auth middleware for testing
	mockAuthMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use mock user ID from header if present, otherwise use default
			userID := r.Header.Get("X-Test-User-ID")
			if userID == "" {
				userID = "00000000-0000-0000-0000-000000000001"
			}
			ctx := middleware.SetUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// Register routes (same as main.go but with mock auth)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Protected endpoints - Rooms
	mux.Handle("/api/rooms", mockAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			roomHandler.CreateRoom(w, r)
		} else if r.Method == http.MethodGet {
			roomHandler.GetRooms(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/rooms/", mockAuthMiddleware(http.HandlerFunc(roomHandler.InviteToRoom)))

	// Protected endpoints - Social
	mux.Handle("/api/follows/", mockAuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			socialHandler.FollowUser(w, r)
		} else if r.Method == http.MethodDelete {
			socialHandler.UnfollowUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/api/me/following", mockAuthMiddleware(http.HandlerFunc(socialHandler.GetFollowing)))
	mux.Handle("/api/users/search", mockAuthMiddleware(http.HandlerFunc(socialHandler.SearchUsers)))

	// Protected endpoints - Sessions
	mux.Handle("/api/sessions", mockAuthMiddleware(http.HandlerFunc(sessionHandler.CreateSession)))
	mux.Handle("/api/sessions/", mockAuthMiddleware(http.HandlerFunc(sessionHandler.GetSession)))

	// Protected endpoints - Voting
	mux.Handle("/api/sessions/{id}/vote", mockAuthMiddleware(http.HandlerFunc(voteHandler.CastVote)))
	mux.Handle("/api/sessions/{id}/matches", mockAuthMiddleware(http.HandlerFunc(matchHandler.GetMatches)))

	// Create test server
	server := httptest.NewServer(mux)

	// Create httpexpect instance
	expect := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewRequireReporter(t),
		Printers: []httpexpect.Printer{
			httpexpect.NewDebugPrinter(t, true),
		},
	})

	return &TestServer{
		Server:     server,
		DB:         testDB,
		Expect:     expect,
		MockUserID: "00000000-0000-0000-0000-000000000001",
	}
}

// Close cleans up the test server and database
func (ts *TestServer) Close() {
	ts.Server.Close()
	ts.DB.Close()
}

// SetMockUserID sets the mock user ID for subsequent requests
func (ts *TestServer) SetMockUserID(userID string) {
	ts.MockUserID = userID
}

// GET creates a GET request with the mock user ID header
func (ts *TestServer) GET(path string) *httpexpect.Request {
	return ts.Expect.GET(path).WithHeader("X-Test-User-ID", ts.MockUserID)
}

// POST creates a POST request with the mock user ID header
func (ts *TestServer) POST(path string) *httpexpect.Request {
	return ts.Expect.POST(path).WithHeader("X-Test-User-ID", ts.MockUserID)
}

// PUT creates a PUT request with the mock user ID header
func (ts *TestServer) PUT(path string) *httpexpect.Request {
	return ts.Expect.PUT(path).WithHeader("X-Test-User-ID", ts.MockUserID)
}

// DELETE creates a DELETE request with the mock user ID header
func (ts *TestServer) DELETE(path string) *httpexpect.Request {
	return ts.Expect.DELETE(path).WithHeader("X-Test-User-ID", ts.MockUserID)
}
