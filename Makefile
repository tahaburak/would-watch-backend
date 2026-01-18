.PHONY: test test-db-up test-db-down test-db-restart test-coverage

# Start test database
test-db-up:
	docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for test database to be ready..."
	@sleep 3

# Stop test database
test-db-down:
	docker-compose -f docker-compose.test.yml down -v

# Restart test database
test-db-restart: test-db-down test-db-up

# Run all tests
test: test-db-up
	@echo "Running tests..."
	TEST_DATABASE_URL="postgres://test_user:test_password@localhost:5433/would_watch_test?sslmode=disable" go test ./... -v

# Run tests with coverage
test-coverage: test-db-up
	@echo "Running tests with coverage..."
	TEST_DATABASE_URL="postgres://test_user:test_password@localhost:5433/would_watch_test?sslmode=disable" go test ./... -coverprofile=coverage.out -covermode=atomic
	@echo "\nCoverage report:"
	@go tool cover -func=coverage.out | grep total
	@echo "\nTo see detailed HTML coverage report, run: go tool cover -html=coverage.out"
