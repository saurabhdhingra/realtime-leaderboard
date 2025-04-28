# Realtime Leaderboard

A Go-based realtime leaderboard system that uses Redis sorted sets for efficient leaderboard management.

## Features

- User authentication (register, login)
- Score submission for different games or activities
- Real-time leaderboard updates
- Global leaderboard across all games
- User ranking queries
- Score history tracking
- Top players report for specific periods
- Built-in metrics and monitoring
- Load testing capability

## Tech Stack

- **Backend**: Go (Gin framework)
- **Database**: Redis (using sorted sets for leaderboards)
- **Authentication**: JWT (JSON Web Tokens)

## Prerequisites

- Go 1.20 or higher
- Redis server

## Getting Started

1. Clone the repository:

```bash
git clone https://github.com/user/realtime-leaderboard.git
cd realtime-leaderboard
```

2. Install dependencies:

```bash
go mod download
```

3. Create a `.env` file in the root directory with the following variables:

```
PORT=8080
REDIS_URL=localhost:6379
JWT_SECRET=your_jwt_secret_key
REDIS_PASSWORD=
REDIS_DB=0
JWT_EXPIRY=24h
```

4. Run the application:

```bash
go run main.go
```

The server will start at http://localhost:8080.

## Using Docker

You can also run the application using Docker:

```bash
# Build and start the application with Redis
docker-compose up

# Or build and start in detached mode
docker-compose up -d
```

## Development Commands (Makefile)

The project includes a Makefile with common development commands:

```bash
# Build the application
make build

# Run the application
make run

# Run tests
make test

# Download dependencies
make deps

# Clean the binary
make clean

# Run with hot-reloading (requires air)
make dev
```

## Testing

### Running Tests

Run the unit tests using:

```bash
go test ./...
```

### API Client

The project includes a sample API client for testing the API. You can run it using:

```bash
go run scripts/api_client.go
```

This will:

1. Register a new user
2. Login with the user's credentials
3. Submit scores to a game
4. Retrieve and display the leaderboard
5. Get the user's rank in the game

### Load Testing

The project includes a load testing tool to benchmark the API under high traffic:

```bash
go run scripts/load_test.go
```

This will:

1. Register multiple concurrent test users
2. Simulate multiple concurrent requests (score submissions and leaderboard retrievals)
3. Collect and display performance metrics like:
   - Total requests processed
   - Success/failure rates
   - Requests per second
   - Average, min, and max response times

You can customize the load test parameters by modifying the constants at the top of the load_test.go file.

## Monitoring

The application includes built-in monitoring capabilities:

- **Metrics API**: Access real-time metrics at `/metrics` endpoint
- **Console Reporting**: Periodic metrics reports in the console (every 5 minutes)
- **Tracked Metrics**:
  - Request counts by method and path
  - Error counts and rates
  - Response times
  - Top endpoints by usage

## API Documentation

For detailed API documentation, see [API.md](API.md).

## API Endpoints

### System Endpoints

- **GET /health**: Health check endpoint
- **GET /metrics**: Get API metrics and statistics

### Authentication

- **POST /api/auth/register**: Register a new user

  - Request Body: `{ "username": "user1", "email": "user1@example.com", "password": "password123" }`

- **POST /api/auth/login**: Login with existing credentials
  - Request Body: `{ "email": "user1@example.com", "password": "password123" }`

### User

- **GET /api/user/profile**: Get the current user's profile
- **GET /api/user/rank/:gameID**: Get the user's rank for a specific game
- **GET /api/user/global-rank**: Get the user's global rank
- **GET /api/user/history/:gameID**: Get the user's score history for a specific game

### Leaderboard

- **POST /api/leaderboard/score**: Submit a new score

  - Request Body: `{ "game_id": "game1", "score": 100 }`

- **GET /api/leaderboard/game/:gameID**: Get the leaderboard for a specific game

  - Query Params: `start=0&count=10`

- **GET /api/leaderboard/global**: Get the global leaderboard

  - Query Params: `start=0&count=10`

- **GET /api/leaderboard/top/:gameID**: Get the top players for a specific period
  - Query Params: `period=day&limit=10` (Valid periods: day, week, month, year)

## Using Redis Sorted Sets for Leaderboards

This project leverages Redis sorted sets to efficiently manage leaderboard data. With sorted sets, Redis provides:

1. O(log(N)) time complexity for adding/updating scores
2. O(log(N)) time complexity for retrieving rank by score
3. O(log(N) + M) time complexity for retrieving a range of ranks

The implementation uses:

- `ZADD` to add or update scores
- `ZREVRANK` to get a user's rank
- `ZREVRANGE` to get a range of leaderboard entries
- `ZINCRBY` to increment scores in the global leaderboard
- `ZRANGEBYSCORE` to get scores within a specific time period

## Project Structure

```
realtime-leaderboard/
├── config/           # Configuration (Redis setup)
├── handlers/         # HTTP handlers for API endpoints
├── middleware/       # Middleware (authentication, metrics)
├── models/           # Data models and business logic
├── scripts/          # Utility scripts (API client, load testing)
├── utils/            # Utility functions (JWT)
├── .env              # Environment variables
├── .gitignore        # Git ignore file
├── API.md            # Detailed API documentation
├── Dockerfile        # Dockerfile for containerization
├── Makefile          # Common development commands
├── README.md         # Project documentation
├── docker-compose.yml # Docker Compose configuration
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
└── main.go           # Application entry point
```

## License

MIT
