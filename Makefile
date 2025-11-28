.PHONY: build test test-docker lint clean coverage

# Build the application
build:
	go build -o bin/ghautodelete ./cmd/ghautodelete

# Run tests in Docker container
test:
	docker run --rm -v $(PWD):/app -w /app golang:1.21 go test -buildvcs=false -v ./...

# Build and run tests via Dockerfile.test
test-docker:
	docker build -f Dockerfile.test -t ghautodelete-test .

# Run linting tools
lint:
	golangci-lint run ./...

# Remove build artifacts
clean:
	rm -rf bin/
	go clean

# Generate test coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
