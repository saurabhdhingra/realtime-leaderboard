# Realtime Leaderboard API Documentation

This document provides detailed information about the Realtime Leaderboard API endpoints, request/response formats, and authentication requirements.

## Base URL

All API endpoints are prefixed with: `/api`

## Authentication

Most endpoints require authentication using JWT (JSON Web Token). To authenticate, include the token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

You can obtain a token by registering or logging in.

## API Endpoints

### System Endpoints

#### Health Check

**Endpoint:** `GET /health`

**Description:** Check if the API is up and running.

**Response (200 OK):**

```json
{
  "status": "ok"
}
```

#### Metrics

**Endpoint:** `GET /metrics`

**Description:** Get API metrics including request counts, error rates, and response times.

**Response (200 OK):**

```json
{
  "total_requests": 1000,
  "total_errors": 50,
  "error_rate": 5.0,
  "requests_by_method": {
    "GET": 700,
    "POST": 300
  },
  "errors_by_method": {
    "GET": 20,
    "POST": 30
  },
  "avg_response_time_ms": {
    "GET": 15.5,
    "POST": 45.2
  },
  "requests_by_path": {
    "GET": {
      "/api/leaderboard/game/:gameID": 350,
      "/api/leaderboard/global": 150,
      "/api/user/profile": 100
    },
    "POST": {
      "/api/auth/login": 150,
      "/api/auth/register": 50,
      "/api/leaderboard/score": 100
    }
  }
}
```

### Authentication

#### Register a New User

**Endpoint:** `POST /api/auth/register`

**Request:**

```json
{
  "username": "user1",
  "email": "user1@example.com",
  "password": "password123"
}
```

**Response (201 Created):**

```json
{
  "message": "User registered successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "user1",
    "email": "user1@example.com"
  }
}
```

#### Login

**Endpoint:** `POST /api/auth/login`

**Request:**

```json
{
  "email": "user1@example.com",
  "password": "password123"
}
```

**Response (200 OK):**

```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "user1",
    "email": "user1@example.com"
  }
}
```

### User

#### Get User Profile

**Endpoint:** `GET /api/user/profile`

**Authentication**: Required

**Response (200 OK):**

```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "user1",
    "email": "user1@example.com",
    "createdAt": "2023-01-01T12:00:00Z"
  }
}
```

#### Get User Rank for a Game

**Endpoint:** `GET /api/user/rank/:gameID`

**Authentication**: Required

**Response (200 OK):**

```json
{
  "ranking": {
    "rank": 1,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "user1",
    "score": 1000
  },
  "game_id": "game1"
}
```

#### Get User Global Rank

**Endpoint:** `GET /api/user/global-rank`

**Authentication**: Required

**Response (200 OK):**

```json
{
  "ranking": {
    "rank": 1,
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "user1",
    "score": 5000
  }
}
```

#### Get User Score History

**Endpoint:** `GET /api/user/history/:gameID`

**Authentication**: Required

**Parameters:**

- `limit` (query, optional): Number of historical scores to retrieve (default: 10)

**Response (200 OK):**

```json
{
  "history": [
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "game_id": "game1",
      "score": 1000,
      "timestamp": "2023-01-02T15:30:00Z"
    },
    {
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "game_id": "game1",
      "score": 950,
      "timestamp": "2023-01-01T12:00:00Z"
    }
  ],
  "game_id": "game1",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "limit": 10
}
```

### Leaderboard

#### Submit a Score

**Endpoint:** `POST /api/leaderboard/score`

**Authentication**: Required

**Request:**

```json
{
  "game_id": "game1",
  "score": 1000
}
```

**Response (200 OK):**

```json
{
  "message": "Score submitted successfully",
  "score": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "game_id": "game1",
    "score": 1000,
    "timestamp": "2023-01-01T12:00:00Z"
  }
}
```

#### Get Game Leaderboard

**Endpoint:** `GET /api/leaderboard/game/:gameID`

**Parameters:**

- `start` (query, optional): Starting index for pagination (default: 0)
- `count` (query, optional): Number of entries to retrieve (default: 10)

**Response (200 OK):**

```json
{
  "leaderboard": [
    {
      "rank": 1,
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "user1",
      "score": 1000
    },
    {
      "rank": 2,
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "username": "user2",
      "score": 950
    }
  ],
  "game_id": "game1",
  "start": 0,
  "count": 10
}
```

#### Get Global Leaderboard

**Endpoint:** `GET /api/leaderboard/global`

**Parameters:**

- `start` (query, optional): Starting index for pagination (default: 0)
- `count` (query, optional): Number of entries to retrieve (default: 10)

**Response (200 OK):**

```json
{
  "leaderboard": [
    {
      "rank": 1,
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "user1",
      "score": 5000
    },
    {
      "rank": 2,
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "username": "user2",
      "score": 4500
    }
  ],
  "start": 0,
  "count": 10
}
```

#### Get Top Players for a Period

**Endpoint:** `GET /api/leaderboard/top/:gameID`

**Parameters:**

- `period` (query, optional): Time period (day, week, month, year) (default: day)
- `limit` (query, optional): Number of entries to retrieve (default: 10)

**Response (200 OK):**

```json
{
  "top_players": [
    {
      "rank": 1,
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "user1",
      "score": 1000
    },
    {
      "rank": 2,
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "username": "user2",
      "score": 950
    }
  ],
  "game_id": "game1",
  "period": "day",
  "start_time": "2023-01-01T00:00:00Z",
  "end_time": "2023-01-02T00:00:00Z",
  "limit": 10
}
```

## Error Responses

All endpoints return standard HTTP status codes. In case of an error, the response body will contain an error message:

```json
{
  "error": "Error message here"
}
```

Common error codes:

- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Invalid or missing authentication
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource already exists (e.g., user email)
- `500 Internal Server Error`: Server-side error

## Rate Limiting

The API implements basic rate limiting to prevent abuse. If you exceed the rate limits, you'll receive a `429 Too Many Requests` response.
