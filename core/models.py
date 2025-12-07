from pydantic import BaseModel, Field
from typing import List, Optional
from datetime import date, datetime
from sqlalchemy.orm import declarative_base, relationship
from sqlalchemy import Column, String, Integer, DateTime, ForeignKey, Index

# --- Domain Constants ---
GLOBAL_LEADERBOARD_KEY = "global_leaderboard"

# --- SQLAlchemy Database Models (Infrastructure Layer) ---
# Base class for declarative models
Base = declarative_base()

class UserDB(Base):
    """SQLAlchemy model for persistent user storage (PostgreSQL)."""
    __tablename__ = "users"
    user_id = Column(String, primary_key=True, index=True)
    username = Column(String, unique=True, index=True, nullable=False)
    hashed_password = Column(String, nullable=False)
    
    # Relationship to score history (optional for ORM use)
    scores = relationship("ScoreHistoryDB", back_populates="user")
    
    # Ensures efficient lookups by username
    __table_args__ = (
        Index('ix_username', 'username'),
    )

class ScoreHistoryDB(Base):
    """SQLAlchemy model for persistent score history (PostgreSQL)."""
    __tablename__ = "score_history"
    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    user_id = Column(String, ForeignKey("users.user_id"), nullable=False)
    game_id = Column(String, index=True, nullable=False)
    score = Column(Integer, nullable=False)
    timestamp = Column(DateTime, default=datetime.utcnow, nullable=False)

    user = relationship("UserDB", back_populates="scores")

    # Indexing for efficient temporal reporting queries
    __table_args__ = (
        Index('ix_game_date_score', 'game_id', 'timestamp', 'score'),
    )

# --- Pydantic Data Models (Core Layer) ---

class User(BaseModel):
    """Represents a registered user in the system."""
    user_id: str
    username: str
    # Removed password hash from the core model

class UserCredentials(BaseModel):
    """Used for registration and login."""
    username: str = Field(..., min_length=3)
    password: str = Field(..., min_length=6)

# --- Data Submission Model ---
class ScoreSubmission(BaseModel):
    """Input model for submitting a new score."""
    game_id: str = Field(..., description="Identifier for the game or activity.")
    score: int = Field(..., ge=0, description="The score achieved in the game.")

# --- Leaderboard Output Models ---
class LeaderboardEntry(BaseModel):
    """An entry for the leaderboard, including rank."""
    rank: int = Field(..., ge=1, description="The user's rank on the leaderboard.")
    user_id: str
    username: str
    score: int
    game_id: str
    
class LeaderboardResponse(BaseModel):
    """The full leaderboard response structure."""
    total_entries: int
    leaderboard: List[LeaderboardEntry]

class UserRankResponse(BaseModel):
    """Response containing a specific user's rank."""
    user_id: str
    username: str
    game_id: str
    score: int
    rank: Optional[int] = Field(..., description="None if user has no score.")

class HistoricalReportEntry(UserRankResponse):
    """
    Model for outputting historical rank data. Used for the "Top Players Report".
    """
    total_scores_submitted: int = 0
    date_range_start: date
    date_range_end: date