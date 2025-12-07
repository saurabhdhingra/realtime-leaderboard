# Realtime Leaderboard Service

This project implements a high-performance, scalable Leaderboard Service using **FastAPI** and a **Clean Architecture** approach. It leverages specialized databases for different needs: **Redis Sorted Sets** for real-time ranking and **PostgreSQL** (via Async SQLAlchemy) for persistent user data and historical score reporting.

## Architecture Overview

The system is structured using a Clean/Hexagonal Architecture pattern, ensuring strict separation of concerns:

1.  **Core Layer (`core/`):** Contains pure domain entities and Pydantic models. No external dependencies.
2.  **Application Layer (`application/`):** Contains interfaces (contracts) for repositories and the central business logic/use cases (Services).
3.  **Infrastructure Layer (`infrastructure/`):** Contains concrete implementations of the repository interfaces using specific technologies (Redis, PostgreSQL).
4.  **API Layer (`api/`):** Contains the entry points (FastAPI endpoints), dependency injection, and security (JWT).

### Data Storage Strategy

| Database | Purpose | Technology Used |
| :--- | :--- | :--- |
| **Redis** | Real-time Leaderboard Ranking | Sorted Sets (ZSET) |
| **PostgreSQL** | User Management, Auth, and Score History | Async SQLAlchemy 2.0 (`asyncpg`) |

## Project Setup and Local Run

### Prerequisites

You need Docker installed to easily run the required databases (PostgreSQL and Redis).

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/saurabhdhingra/realtime-leaderboard/
    cd leaderboard_app
    ```

2.  **Create a virtual environment and install dependencies:**

    ```bash
    # Assuming Python 3.10+
    python -m venv venv
    source venv/bin/activate
    pip install -r requirements.txt
    ```

3.  **Set up Environment Variables:**

    Create a file named `.env` in the project root (`leaderboard_app/.env`) and populate it with your database connection details and JWT secret key:

    ```env
    # --- PostgreSQL Configuration ---
    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=mysecretpassword
    POSTGRES_HOST=localhost
    POSTGRES_PORT=5432
    POSTGRES_DB=leaderboard_db

    # --- Redis Configuration ---
    REDIS_HOST=localhost
    REDIS_PORT=6379

    # --- JWT Security (Must be strong and unique) ---
    SECRET_KEY="SUPER_SECRET_KEY_REPLACE_ME_IN_PRODUCTION" 
    ```

4.  **Start the Databases with Docker:**

    You can use a `docker-compose.yml` (not included, but assumed) or run them manually:

    ```bash
    # Start PostgreSQL
    docker run --name leaderboard-postgres -e POSTGRES_PASSWORD=mysecretpassword -p 5432:5432 -d postgres

    # Start Redis
    docker run --name leaderboard-redis -p 6379:6379 -d redis
    ```

### Running the Application

Execute the `main.py` file using `uvicorn`:

```bash
uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload
```

The application will be available at `http://localhost:8000`.

## Authentication Flow (JWT)

All secure endpoints (submitting scores, getting user rank, reports) require a JWT access token.

1.  **Register:** Send credentials to `POST /v1/auth/register`.
2.  **Login:** Send credentials to `POST /v1/auth/token`. This returns a JWT `access_token`.
3.  **Access:** Include the token in subsequent request headers: `Authorization: Bearer [access_token]`.

## Key Endpoints

| Method | Path | Description | Authentication | Data Store |
| :--- | :--- | :--- | :--- | :--- |
| `POST` | `/v1/auth/register` | Creates a new user record in PostgreSQL. | Public | PostgreSQL |
| `POST` | `/v1/auth/token` | Authenticates and returns a JWT access token. | Public | PostgreSQL |
| `POST` | `/v1/scores/submit` | Updates the highest score in Redis and logs the event in PostgreSQL History. | Required | Redis & PostgreSQL |
| `GET` | `/v1/leaderboard/global` | Retrieves the top players, paginated. | Public | Redis |
| `GET` | `/v1/leaderboard/my_rank` | Retrieves the logged-in user's current rank and score. | Required | Redis |
| `GET` | `/v1/reports/top_players` | Generates a report on top players based on the highest score achieved between a `start_date` and `end_date`. | Required | PostgreSQL |

## Key Production Features

  * **Asynchronous I/O:** Utilizes `asyncio`, `FastAPI`, `asyncpg`, and `redis.asyncio` for non-blocking operations, maximizing API throughput.
  * **Secure Authentication:** Implements JWT for stateless security and uses `passlib` (Bcrypt) for secure password hashing.
  * **Optimized Data Access:** The ranking endpoint reads directly from Redis, providing extremely low-latency responses, while reporting endpoints leverage PostgreSQL's power for complex historical data aggregation.
  * **Dependency Injection:** Uses FastAPI's `Depends` system to inject configured Repository instances (`RedisLeaderboardRepository`, `PostgresUserRepository`) and the main `LeaderboardService` instance, promoting testability and configuration flexibility.

## Acknowledgement
https://roadmap.sh/projects/realtime-leaderboard-system
