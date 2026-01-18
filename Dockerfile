# Build Stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/api/main.go

# Run Stage
FROM alpine:latest

WORKDIR /app

# Install certificates for external APIs (Supabase/TMDB)
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .
# Copy migrations if needed (or run them separately)
COPY db/schema.sql ./db/schema.sql

# Expose port (must match PORT env var)
EXPOSE 8080

# Run the binary
CMD ["./main"]
