from application.interfaces import IUserRepository
from core.models import User, UserDB
from typing import List, Optional, Dict
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select, func, update

class PostgresUserRepository(IUserRepository):
    """
    Production implementation of IUserRepository using Async SQLAlchemy (PostgreSQL).
    """
    def __init__(self, db_session: AsyncSession):
        self.session = db_session

    async def get_user_by_id(self, user_id: str) -> Optional[User]:
        stmt = select(UserDB).where(UserDB.user_id == user_id)
        result = await self.session.execute(stmt)
        user_db = result.scalars().first()
        return User.model_validate(user_db) if user_db else None

    async def get_user_by_username(self, username: str) -> Optional[User]:
        stmt = select(UserDB).where(UserDB.username == username)
        result = await self.session.execute(stmt)
        user_db = result.scalars().first()
        return User.model_validate(user_db) if user_db else None

    async def create_user(self, user: User, hashed_password: str) -> User:
        user_db = UserDB(
            user_id=user.user_id,
            username=user.username,
            hashed_password=hashed_password
        )
        self.session.add(user_db)
        await self.session.commit()
        await self.session.refresh(user_db)
        return User.model_validate(user_db)

    async def get_username_map(self, user_ids: List[str]) -> Dict[str, str]:
        if not user_ids:
            return {}
            
        stmt = select(UserDB.user_id, UserDB.username).where(UserDB.user_id.in_(user_ids))
        result = await self.session.execute(stmt)
        
        # Returns a map of {user_id: username}
        return {uid: uname for uid, uname in result.all()}
        
    async def get_hashed_password(self, username: str) -> Optional[str]:
        stmt = select(UserDB.hashed_password).where(UserDB.username == username)
        result = await self.session.execute(stmt)
        return result.scalars().first()