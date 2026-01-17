# Auth Middleware

HTTP middleware for validating Supabase JWT tokens and protecting API endpoints.

## Features

- Validates JWT tokens from Authorization headers
- Verifies token signature using HMAC-SHA256
- Extracts user ID from token claims
- Injects user ID into request context
- Comprehensive error handling

## Usage

### Basic Setup

```go
import (
    "github.com/burak/would_watch/backend/internal/middleware"
    "net/http"
)

// Create the auth middleware with your Supabase JWT secret
authMiddleware := middleware.AuthMiddleware("your-jwt-secret")

// Wrap your protected handlers
protectedHandler := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Get user ID from context
    userID, ok := middleware.GetUserID(r.Context())
    if !ok {
        http.Error(w, "User ID not found", http.StatusInternalServerError)
        return
    }

    // Use userID in your handler
    w.Write([]byte("Hello, " + userID))
}))

http.Handle("/api/protected", protectedHandler)
```

### Making Authenticated Requests

Clients must include a valid JWT token in the Authorization header:

```bash
curl -H "Authorization: Bearer <your-jwt-token>" \
     http://localhost:8080/api/protected
```

## How It Works

1. **Token Extraction**: The middleware extracts the JWT token from the `Authorization: Bearer <token>` header
2. **Token Validation**:
   - Verifies the token signature using the provided JWT secret
   - Checks token expiration
   - Validates the signing method (HMAC-SHA256)
3. **User ID Extraction**: Extracts the user ID from the `sub` claim in the JWT
4. **Context Injection**: Injects the user ID into the request context for use by handlers

## Error Responses

The middleware returns HTTP 401 Unauthorized for:
- Missing Authorization header
- Invalid header format (not "Bearer <token>")
- Invalid or expired tokens
- Tokens signed with wrong secret
- Tokens missing the `sub` claim
- Tokens using unexpected signing methods

## Configuration

The middleware requires the Supabase JWT secret, which can be found in your Supabase project settings:

1. Go to your Supabase project dashboard
2. Navigate to Settings > API
3. Copy the JWT Secret

Add it to your environment variables:
```bash
export SUPABASE_JWT_SECRET="your-jwt-secret-here"
```

## Helper Functions

### GetUserID

Extract the user ID from a request context:

```go
userID, ok := middleware.GetUserID(r.Context())
if !ok {
    // User ID not found in context
    // This shouldn't happen if the middleware is applied correctly
}
```

## Testing

Run tests with:
```bash
go test ./internal/middleware/... -v
```

The test suite covers:
- Missing authorization headers
- Invalid header formats
- Invalid tokens
- Expired tokens
- Wrong JWT secrets
- Missing claims
- Valid token flows
