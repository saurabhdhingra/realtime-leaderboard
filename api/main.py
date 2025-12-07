import os
import asyncio
from fastapi import FastAPI, HTTPException, status
from redis.asyncio import Redis
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker
from dotenv import load_dotenv

from core.models import Base
from infrastructure.redis_repository import RedisLeaderboardRepository
from application.services import LeaderboardService
from api.endpoints import router as leaderboard_router
from api.dependencies import get_leaderboard_service, async_session_maker, get_current_user 

load_dotenv() 

POSTGRES_USER = os.getenv("POSTGRES_USER", "postgres")
POSTGRES_PASSWORD = os.getenv("POSTGRES_PASSWORD", "password")
POSTGRES_HOST = os.getenv("POSTGRES_HOST", "localhost")
POSTGRES_PORT = os.getenv("POSTGRES_PORT", "5432")
POSTGRES_DB = os.getenv("POSTGRES_DB", "leaderboard_db")

REDIS_HOST = os.getenv("REDIS_HOST", "localhost")
REDIS_PORT = int(os.getenv("REDIS_PORT", 6379))

DATABASE_URL = f"postgresql+asyncpg://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}"

redis_repo_instance: Optional[RedisLeaderboardRepository] = None


async def create_db_and_tables():
    """Initializes the database engine and creates tables if they don't exist."""
    print("Attempting to connect to PostgreSQL...")

    async_engine = create_async_engine(DATABASE_URL, echo=True) 

    global async_session_maker
    async_session_maker = async_sessionmaker(
        async_engine, 
        autoflush=False, 
        expire_on_commit=False
    )

    async with async_engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    
    print("PostgreSQL tables created or verified.")


def create_app() -> FastAPI:
    app = FastAPI(
        title="Production Leaderboard Service (Clean Architecture)",
        description="FastAPI orchestrating Redis (Real-Time) and PostgreSQL (History/Auth).",
        version="1.0.0"
    )

    @app.on_event("startup")
    async def startup_event():
        try:
            redis_client = Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)
            await redis_client.ping()
            print(f"Successfully connected to Redis at {REDIS_HOST}:{REDIS_PORT}")
            
            global redis_repo_instance
            redis_repo_instance = RedisLeaderboardRepository(redis_client)
            
        except Exception as e:
            print(f"CRITICAL: Failed to connect to Redis: {e}")
            
        await create_db_and_tables()

    @app.on_event("shutdown")
    async def shutdown_event():
        if redis_repo_instance and redis_repo_instance.redis:
            await redis_repo_instance.redis.close()
            print("Redis client closed.")

    def get_redis_repo_override():
        if redis_repo_instance is None:
             raise HTTPException(status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail="Service Unavailable: Redis connection failed.")
        return redis_repo_instance

    app.dependency_overrides[type(redis_repo_instance)] = get_redis_repo_override

    app.include_router(leaderboard_router, prefix="/v1")

    return app

app = create_app()

if __name__ == "__main__":
    print("Run using: uvicorn api.main:app --host 0.0.0.0 --port 8000 --reload")