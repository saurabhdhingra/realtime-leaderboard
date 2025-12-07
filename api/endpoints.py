from fastapi import APIRouter, Depends, HTTPException, status
from application.services import LeaderboardService
from api.dependencies import get_leaderboard_service, get_current_user, create_access_token, pwd_context
from core.models import (
    ScoreSubmission, 
    LeaderboardResponse, 
    UserRankResponse, 
    UserCredentials, 
    User,
    GLOBAL_LEADERBOARD_KEY,
    HistoricalReportEntry
)
from typing import Dict, Any, List
from datetime import date, timedelta
from pydantic import ValidationError

router = APIRouter()


@router.post("/auth/register", response_model=User)
async def register_user_endpoint(
    creds: UserCredentials, 
    service: LeaderboardService = Depends(get_leaderboard_service)
):
    try:
        hashed_password = pwd_context.hash(creds.password)
        
        user = await service.register_user(creds, hashed_password)
        return user
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))

@router.post("/auth/token", response_model=Dict[str, str])
async def login_for_access_token(
    creds: UserCredentials, 
    service: LeaderboardService = Depends(get_leaderboard_service)
):
    """
    Standard OAuth2 endpoint. Authenticates user and returns JWT access token.
    """
    try:
        user = await service.login_user(creds)

        access_token_expires = timedelta(minutes=30) 
        access_token = create_access_token(
            data={"user_id": user.user_id}, expires_delta=access_token_expires
        )
        
        return {"access_token": access_token, "token_type": "bearer", "username": user.username}
        
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, 
            detail=str(e), 
            headers={"WWW-Authenticate": "Bearer"}
        )


@router.post("/scores/submit", response_model=Dict[str, Any])
async def submit_score_endpoint(
    submission: ScoreSubmission,
    current_user: User = Depends(get_current_user),
    service: LeaderboardService = Depends(get_leaderboard_service)
):
    """Submits a score to Redis (real-time) and PostgreSQL (history)."""
    score = await service.submit_new_score(
        user_id=current_user.user_id,
        game_id=submission.game_id,
        submission=submission
    )
    return {
        "user_id": current_user.user_id,
        "username": current_user.username,
        "score_saved": score,
        "message": "Score submitted successfully."
    }

@router.get("/leaderboard/global", response_model=LeaderboardResponse)
async def get_global_leaderboard_endpoint(
    page: int = 1, 
    page_size: int = 20,
    service: LeaderboardService = Depends(get_leaderboard_service)
):
    """Retrieves the global leaderboard (Real-time from Redis)."""
    return await service.get_leaderboard(
        game_id=GLOBAL_LEADERBOARD_KEY, 
        page=page, 
        page_size=page_size
    )

@router.get("/leaderboard/my_rank", response_model=UserRankResponse)
async def get_user_rank_endpoint(
    game_id: str = GLOBAL_LEADERBOARD_KEY,
    current_user: User = Depends(get_current_user), 
    service: LeaderboardService = Depends(get_leaderboard_service)
):
    """Retrieves the logged-in user's current rank (Real-time from Redis)."""
    return await service.get_user_rank(
        user_id=current_user.user_id,
        game_id=game_id
    )
    
# --- Historical Reporting Endpoint (Requires Auth) ---

@router.get("/reports/top_players", response_model=List[HistoricalReportEntry])
async def get_top_players_report_endpoint(
    game_id: str,
    start_date: date,
    end_date: date,
    limit: int = 10,
    current_user: User = Depends(get_current_user),
    service: LeaderboardService = Depends(get_leaderboard_service)
):
    """
    Generates a report of top players for a specific period using the
    persistent score history (PostgreSQL).
    """
    try:
        if start_date > end_date:
            raise HTTPException(status_code=400, detail="Start date cannot be after end date.")
            
        return await service.get_historical_report(
            game_id=game_id,
            start_date=start_date,
            end_date=end_date,
            limit=limit
        )
    except ValidationError as e:
        raise HTTPException(status_code=422, detail=e.errors())
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))