package middleware

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for storing user ID
	UserIDKey ContextKey = "userID"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// AuthMiddleware creates a middleware that validates Supabase JWT tokens using JWT secret
func AuthMiddleware(supabaseURL string, jwtSecret string) func(http.Handler) http.Handler {
	// Decode base64-encoded JWT secret
	fmt.Printf("🔐 [AuthMiddleware] JWT secret (first 20 chars): %s...\n", jwtSecret[:min(20, len(jwtSecret))])
	secretKey, err := base64.StdEncoding.DecodeString(jwtSecret)
	if err != nil {
		fmt.Printf("❌ [AuthMiddleware] Failed to decode JWT secret: %v\n", err)
		// Fallback to using the secret as-is if it's not base64
		secretKey = []byte(jwtSecret)
	} else {
		fmt.Printf("✅ [AuthMiddleware] JWT secret decoded successfully, %d bytes\n", len(secretKey))
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

			fmt.Printf("🔐 [AuthMiddleware] Validating token (first 50 chars): %s...\n", tokenString[:min(50, len(tokenString))])
			fmt.Printf("🔐 [AuthMiddleware] Using secret key of %d bytes\n", len(secretKey))

			// Parse and validate the JWT token using the JWT secret
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Verify the signing method is HMAC
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					fmt.Printf("❌ [AuthMiddleware] Unexpected signing method: %v\n", token.Header["alg"])
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				fmt.Printf("✅ [AuthMiddleware] Signing method is HMAC: %v\n", token.Method.Alg())
				return secretKey, nil
			})

			if err != nil {
				fmt.Printf("❌ [AuthMiddleware] Token validation error: %v\n", err) // Debug log
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
