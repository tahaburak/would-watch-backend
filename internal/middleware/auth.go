package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for storing user ID
	UserIDKey ContextKey = "userID"
)

// AuthMiddleware creates a middleware that validates Supabase JWT tokens using JWKS
func AuthMiddleware(supabaseURL string) func(http.Handler) http.Handler {
	// Construct JWKS URL
	jwksURL := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", supabaseURL)

	// Create JWKS keyfunc
	// In a real app, you might want to handle initialization error better or retry
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		fmt.Printf("Failed to create JWKS from resource at the given URL.\nError: %s\n", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing authorization header", http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// If JWKS failed to initialize, return error
			if jwks == nil {
				http.Error(w, "Authentication service unavailable", http.StatusServiceUnavailable)
				return
			}

			// Parse and validate the JWT token using the JWKS keyfunc
			token, err := jwt.Parse(tokenString, jwks.Keyfunc)

			if err != nil {
				fmt.Printf("Token validation error: %v\n", err) // Debug log
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Extract user ID from claims
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			// Supabase JWT tokens contain the user ID in the "sub" claim
			userID, ok := claims["sub"].(string)
			if !ok || userID == "" {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}

			// Inject user ID into context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// SetUserID sets the user ID in the context (for testing)
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
