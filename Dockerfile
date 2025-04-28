FROM golang:1.20-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./
COPY go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /realtime-leaderboard

# Use a small image for the final image
FROM alpine:latest

# Add CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /realtime-leaderboard /realtime-leaderboard

# Copy the .env file
COPY .env /.env

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["/realtime-leaderboard"] 