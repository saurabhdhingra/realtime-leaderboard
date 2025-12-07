from abc import ABC, abstractmethod
from typing import List, Optional, Dict
from core.models import User, ScoreSubmission, HistoricalReportEntry
from datetime import date

# --- User Repository Contract (PostgreSQL) ---

class IUserRepository(ABC):
    """
    Contract for user data management, backed by a persistent DB (e.g., PostgreSQL).
    Handles user creation, retrieval, and authentication details.
    """
    @abstractmethod
    async def get_user_by_id(self, user_id: str) -> Optional[User]:
        """Retrieve user details by ID."""
        pass
        
    @abstractmethod
    async def get_user_by_username(self, username: str) -> Optional[User]:
        """Retrieve user details by username."""
        pass

    @abstractmethod
    async def create_user(self, user: User, hashed_password: str) -> User:
        """Create a new user record with a hashed password."""
        pass

    @abstractmethod
    async def get_username_map(self, user_ids: List[str]) -> Dict[str, str]:
        """Fetch usernames for a list of IDs efficiently (e.g., batch query)."""
        pass
        
    @abstractmethod
    async def get_hashed_password(self, username: str) -> Optional[str]:
        """Retrieve the stored password hash for authentication verification."""
        pass

# --- Score History Repository Contract (PostgreSQL) ---

class IScoreHistoryRepository(ABC):
    """
    Contract for persisting and querying all score submissions for temporal reports.
    """
    @abstractmethod
    async def log_score_submission(self, user_id: str, submission: ScoreSubmission):
        """Persist every score submission event."""
        pass

    @abstractmethod
    async def get_top_players_report(self, game_id: str, start_date: date, end_date: date, limit: int = 10) -> List[HistoricalReportEntry]:
        """
        Generates a report of the top players based on the highest score achieved
        within the specified time period, or aggregated score. (PostgreSQL Query)
        """
        pass

# --- Leaderboard Repository Contract (Redis) ---

class ILeaderboardRepository(ABC):
    """
    Contract for real-time ranking data storage, backed by Redis Sorted Sets.
    """
    @abstractmethod
    async def submit_score(self, user_id: str, game_id: str, score: int) -> int:
        """
        Submits a score to the specific game leaderboard.
        Returns the score saved (the current highest score).
        """
        pass

    @abstractmethod
    async def get_leaderboard_page(self, game_id: str, start: int, end: int) -> List[Dict]:
        """Retrieve a specific range of the leaderboard."""
        pass

    @abstractmethod
    async def get_total_entries(self, game_id: str) -> int:
        """Get the total number of entries in a leaderboard."""
        pass

    @abstractmethod
    async def get_user_rank_and_score(self, user_id: str, game_id: str) -> Optional[Dict]:
        """Get a user's rank and score for a specific leaderboard."""
        pass