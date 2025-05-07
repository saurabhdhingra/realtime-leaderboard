.PHONY: build run clean test

build:
	go build -o realtime-leaderboard

run:
	go run main.go

clean:
	rm -f realtime-leaderboard

test:
	go test ./...

deps:
	go mod download

dev:
	air 