from redis.asyncio import Redis 
from application.interfaces import ILeaderboardRepository
from core.models import GLOBAL_LEADERBOARD_KEY
from typing import List, Optional, Dict

MOCK_USERS: Dict[str, Dict] = {} 

class RedisLeaderboardRepository(ILeaderboardRepository):
    """
    Implementation of the ILeaderboardRepository using Redis Sorted Sets (ZSET).
    Uses the asynchronous Redis client for non-blocking operations.
    """
    def __init__(self, redis_client: Redis):
        self.redis = redis_client

    async def _get_zset_key(self, game_id: str) -> str:
        """In this design, we use one global leaderboard for simplicity."""
        return GLOBAL_LEADERBOARD_KEY 

    async def submit_score(self, user_id: str, game_id: str, score: int) -> int:
        """
        Adds or updates the user's score. Uses GT (Greater Than) to ensure 
        only higher scores are saved, maintaining the "highest score" paradigm.
        """
        key = await self._get_zset_key(game_id)
        
        saved_score = await self.redis.zadd(
            key,
            {user_id: score},
            gt=True 
        )

        if saved_score == 0:
            current_score = await self.redis.zscore(key, user_id)
            return int(current_score) if current_score is not None else score
        
        return score

    async def get_leaderboard_page(self, game_id: str, start: int, end: int) -> List[Dict]:
        """
        Retrieves a range of users and scores from the ZSET (ZREVRANGE for descending score).
        """
        key = await self._get_zset_key(game_id)
        
        raw_results = await self.redis.zrevrange( 
            key, 
            start, 
            end, 
            withscores=True
        )
        
        entries = []
        for user_id_bytes, score_float in raw_results:
            entries.append({
                'user_id': user_id_bytes.decode('utf-8'),
                'score': int(score_float)
            })
            
        return entries

    async def get_total_entries(self, game_id: str) -> int:
        """Get the total number of members in the ZSET."""
        key = await self._get_zset_key(game_id)
        return await self.redis.zcard(key)

    async def get_user_rank_and_score(self, user_id: str, game_id: str) -> Optional[Dict]:
        """
        Retrieves the 0-based rank (ZREVRANK) and score (ZSCORE) of a user.
        """
        key = await self._get_zset_key(game_id)

        rank = await self.redis.zrevrank(key, user_id)
        
        if rank is None:
            return None
            
        score = await self.redis.zscore(key, user_id)
        
        return {
            'rank': rank, 
            'score': int(score) if score is not None else 0
        }