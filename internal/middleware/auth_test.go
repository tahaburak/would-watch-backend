package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

// createTestToken creates a valid JWT token for testing
func createTestToken(userID string, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	middleware := AuthMiddleware(testSecret)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if !contains(rr.Body.String(), "Missing authorization header") {
		t.Errorf("expected error message about missing header, got: %s", rr.Body.String())
	}
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	middleware := AuthMiddleware(testSecret)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "some-token"},
		{"Wrong prefix", "Basic some-token"},
		{"Only Bearer", "Bearer"},
		{"Too many parts", "Bearer token extra"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
			}
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	middleware := AuthMiddleware(testSecret)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if !contains(rr.Body.String(), "Invalid token") {
		t.Errorf("expected error message about invalid token, got: %s", rr.Body.String())
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	middleware := AuthMiddleware(testSecret)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create token with different secret
	wrongToken := createTestToken("user-123", "wrong-secret")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+wrongToken)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	middleware := AuthMiddleware(testSecret)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create expired token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAuthMiddleware_MissingSubClaim(t *testing.T) {
	middleware := AuthMiddleware(testSecret)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create token without sub claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	if !contains(rr.Body.String(), "Invalid user ID") {
		t.Errorf("expected error message about invalid user ID, got: %s", rr.Body.String())
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	middleware := AuthMiddleware(testSecret)

	expectedUserID := "user-123-abc"
	var capturedUserID string

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			t.Error("expected to find user ID in context")
		}
		capturedUserID = userID
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	validToken := createTestToken(expectedUserID, testSecret)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if capturedUserID != expectedUserID {
		t.Errorf("expected user ID %s, got %s", expectedUserID, capturedUserID)
	}

	if rr.Body.String() != "success" {
		t.Errorf("expected body 'success', got: %s", rr.Body.String())
	}
}

func TestAuthMiddleware_WrongSigningMethod(t *testing.T) {
	middleware := AuthMiddleware(testSecret)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Create token with RSA signing method instead of HMAC
	// Note: This will fail during parsing since we expect HMAC
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestGetUserID_NotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	_, ok := GetUserID(req.Context())

	if ok {
		t.Error("expected user ID not to be found in empty context")
	}
}

func TestGetUserID_Found(t *testing.T) {
	middleware := AuthMiddleware(testSecret)
	expectedUserID := "user-456"
	var capturedUserID string
	var foundInContext bool

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, foundInContext = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	validToken := createTestToken(expectedUserID, testSecret)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !foundInContext {
		t.Error("expected to find user ID in context")
	}

	if capturedUserID != expectedUserID {
		t.Errorf("expected user ID %s, got %s", expectedUserID, capturedUserID)
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
