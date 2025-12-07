from pydantic import BaseModel, Field
from typing import List, Optional

# --- Domain Constants ---
GLOBAL_LEADERBOARD_KEY = "global_leaderboard"

# --- User & Auth Models (Mocked) ---
class User(BaseModel):
    """Represents a registered user in the system."""
    user_id: str
    username: str

class UserCredentials(BaseModel):
    """Used for mock registration and login."""
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