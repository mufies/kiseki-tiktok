"""Feed router — GET /feed/{user_id}, GET /trending, GET /health."""
from __future__ import annotations

from fastapi import APIRouter, Depends, Query, Request

from app.models import FeedResponse, HealthResponse, TrendingResponse, VideoResult
from app.services.feed_service import FeedService

router = APIRouter()


def _get_feed_service(request: Request) -> FeedService:
    return request.app.state.feed_service


@router.get("/feed/{user_id}", response_model=FeedResponse)
async def get_feed(
    user_id: str,
    limit: int = Query(default=20, ge=1, le=100),
    feed_service: FeedService = Depends(_get_feed_service),
) -> FeedResponse:
    phase, videos = await feed_service.get_feed(user_id, limit)
    return FeedResponse(user_id=user_id, phase=phase, videos=videos)


@router.get("/trending", response_model=TrendingResponse)
async def get_trending(
    limit: int = Query(default=20, ge=1, le=100),
    feed_service: FeedService = Depends(_get_feed_service),
) -> TrendingResponse:
    videos = await feed_service.get_trending(limit)
    return TrendingResponse(videos=videos)


@router.get("/health", response_model=HealthResponse)
async def health(request: Request) -> HealthResponse:
    state = request.app.state

    pg_ok    = False
    redis_ok = False

    try:
        await state.db_pool.fetchval("SELECT 1")
        pg_ok = True
    except Exception:
        pass

    try:
        await state.redis_client.ping()
        redis_ok = True
    except Exception:
        pass

    overall = "healthy" if (pg_ok and redis_ok) else "degraded"

    return HealthResponse(
        status=overall,
        postgres="ok" if pg_ok else "unreachable",
        redis="ok" if redis_ok else "unreachable",
    )
