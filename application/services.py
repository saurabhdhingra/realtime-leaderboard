from application.interfaces import ILeaderboardRepository, IUserRepository, IScoreHistoryRepository, HistoricalReportEntry
from core.models import ScoreSubmission, LeaderboardResponse, UserRankResponse, UserCredentials, User
from typing import List
from datetime import date
from uuid import uuid4


def hash_password(password: str) -> str:
    """Production Placeholder: Mocks password hashing."""
    return pwd_context.hash(password)

def verify_password(plain_password: str, hashed_password: str) -> bool:
    """Production Placeholder: Mocks password verification."""
    return pwd_context.verify(plain_password, hashed_password)


class LeaderboardService:
    """
    Handles the core business logic, orchestrating calls to the 
    User (PostgreSQL), Leaderboard (Redis), and Score History (PostgreSQL) repositories.
    """
    def __init__(self, 
                 leaderboard_repo: ILeaderboardRepository,
                 user_repo: IUserRepository,
                 history_repo: IScoreHistoryRepository):
        
        self.leaderboard_repo = leaderboard_repo
        self.user_repo = user_repo
        self.history_repo = history_repo


    async def register_user(self, creds: UserCredentials) -> User:
        """
        Registers a new user using the UserRepository (PostgreSQL).
        """
        existing_user = await self.user_repo.get_user_by_username(creds.username)
        if existing_user:
            raise ValueError("Username already exists.")
        
        hashed_password = hash_password(creds.password)
        new_user = User(user_id=str(uuid4()), username=creds.username)
        
        return await self.user_repo.create_user(new_user, hashed_password)

    async def login_user(self, creds: UserCredentials) -> User:
        """
        Logs in a user by verifying credentials against the UserRepository.
        """
        user = await self.user_repo.get_user_by_username(creds.username)
        if not user:
            raise ValueError("Invalid username or password.")
        
        stored_hash = await self.user_repo.get_hashed_password(user.username)
        
        if not stored_hash or not verify_password(creds.password, stored_hash):
             raise ValueError("Invalid username or password.")
        
        return user


    async def submit_new_score(self, user_id: str, game_id: str, submission: ScoreSubmission) -> int:
        """
        Submits a score, updating both the real-time leaderboard (Redis) 
        and the persistent history (PostgreSQL).
        """
        await self.history_repo.log_score_submission(user_id, submission)

        saved_score = await self.leaderboard_repo.submit_score(
            user_id=user_id,
            game_id=game_id,
            score=submission.score
        )
        return saved_score


    async def get_leaderboard(self, game_id: str, page: int = 1, page_size: int = 20) -> LeaderboardResponse:
        """Retrieve a paginated, ranked list of users for a given game from Redis."""
        
        start = (page - 1) * page_size
        end = start + page_size - 1

        raw_entries = await self.leaderboard_repo.get_leaderboard_page(game_id, start, end)
        total_entries = await self.leaderboard_repo.get_total_entries(game_id)
        
        if not raw_entries:
            return LeaderboardResponse(total_entries=0, leaderboard=[])

        user_ids = [entry['user_id'] for entry in raw_entries]
        
        username_map = await self.user_repo.get_username_map(user_ids)

        leaderboard_data: List[LeaderboardEntry] = []
        for i, entry in enumerate(raw_entries):
            user_id = entry['user_id']
            leaderboard_data.append(
                LeaderboardEntry(
                    rank=start + i + 1,
                    user_id=user_id,
                    username=username_map.get(user_id, "Unknown User"),
                    score=entry['score'],
                    game_id=game_id
                )
            )

        return LeaderboardResponse(
            total_entries=total_entries,
            leaderboard=leaderboard_data
        )

    async def get_user_rank(self, user_id: str, game_id: str) -> UserRankResponse:
        """Retrieve a specific user's rank and score from Redis."""
        
        user_data = await self.leaderboard_repo.get_user_rank_and_score(user_id, game_id)

        username_map = await self.user_repo.get_username_map([user_id])
        username = username_map.get(user_id, "Unknown User")
        
        if not user_data:
            return UserRankResponse(
                user_id=user_id,
                username=username,
                game_id=game_id,
                score=0,
                rank=None
            )

        rank = user_data['rank'] + 1
        score = user_data['score']

        return UserRankResponse(
            user_id=user_id,
            username=username,
            game_id=game_id,
            score=score,
            rank=rank
        )
        
    
    async def get_historical_report(
        self, 
        game_id: str, 
        start_date: date, 
        end_date: date, 
        limit: int = 10
    ) -> List[HistoricalReportEntry]:
        """
        Generates a report of top players for a specific period using the
        persistent score history (PostgreSQL).
        """
        return await self.history_repo.get_top_players_report(
            game_id=game_id,
            start_date=start_date,
            end_date=end_date,
            limit=limit
        )