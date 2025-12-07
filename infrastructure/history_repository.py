from application.interfaces import IScoreHistoryRepository, HistoricalReportEntry
from core.models import ScoreSubmission, ScoreHistoryDB
from typing import List
from datetime import date
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func, extract, and_
from sqlalchemy.sql.expression import alias
from infrastructure.user_repository_postgres import PostgresUserRepository # For dependency access

class PostgresScoreHistoryRepository(IScoreHistoryRepository):
    """
    Production implementation of IScoreHistoryRepository using Async SQLAlchemy (PostgreSQL).
    This handles complex aggregation queries for temporal reports.
    """
    def __init__(self, db_session: AsyncSession):
        self.session = db_session

    async def log_score_submission(self, user_id: str, submission: ScoreSubmission):
        """Persist every score submission event."""
        history_db = ScoreHistoryDB(
            user_id=user_id,
            game_id=submission.game_id,
            score=submission.score
        )
        self.session.add(history_db)
        await self.session.flush()

    async def get_top_players_report(self, game_id: str, start_date: date, end_date: date, limit: int = 10) -> List[HistoricalReportEntry]:
        """
        Generates a report: top players based on their highest single score
        achieved within the date range for a specific game.
        """

        subquery = (
            select(
                ScoreHistoryDB.user_id,
                func.max(ScoreHistoryDB.score).label("highest_score"),
                func.count(ScoreHistoryDB.id).label("total_scores_submitted")
            )
            .where(
                and_(
                    ScoreHistoryDB.game_id == game_id,
                    ScoreHistoryDB.timestamp.between(start_date, end_date)
                )
            )
            .group_by(ScoreHistoryDB.user_id)
            .order_by(func.max(ScoreHistoryDB.score).desc())
            .limit(limit)
            .subquery()
        )
        
        stmt = select(
            subquery.c.user_id,
            subquery.c.highest_score,
            subquery.c.total_scores_submitted
        )
        
        result = await self.session.execute(stmt)
        raw_report_data = result.all()
        
        user_ids = [row.user_id for row in raw_report_data]

        username_map = {uid: f"User_{uid[:4]}" for uid in user_ids} 


        report: List[HistoricalReportEntry] = []
        for i, row in enumerate(raw_report_data):
            report.append(HistoricalReportEntry(
                rank=i + 1,
                user_id=row.user_id,
                username=username_map.get(row.user_id, "Unknown User"),
                game_id=game_id,
                score=row.highest_score,
                total_scores_submitted=row.total_scores_submitted,
                date_range_start=start_date,
                date_range_end=end_date
            ))
            
        return report