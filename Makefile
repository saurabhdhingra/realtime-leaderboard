.PHONY: build run clean test

# Build the application
build:
	go build -o realtime-leaderboard

# Run the application
run:
	go run main.go

# Clean the binary
clean:
	rm -f realtime-leaderboard

# Run tests
test:
	go test ./...

# Download dependencies
deps:
	go mod download

# Run with hot reloading (requires air: https://github.com/cosmtrek/air)
dev:
	air 