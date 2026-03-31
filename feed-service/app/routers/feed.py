"""Feed router — GET /feed/{user_id}, GET /trending, GET /health."""
from __future__ import annotations

from fastapi import APIRouter, Depends, Query, Request

from app.models import FeedResponse, HealthResponse, TrendingResponse
from app.protocols.services import IFeedProvider, ITrendingProvider

router = APIRouter()


def _get_feed_service(request: Request) -> IFeedProvider:
    return request.app.state.feed_service


def _get_trending_service(request: Request) -> ITrendingProvider:
    return request.app.state.trending_service


@router.get("/feed/{user_id}", response_model=FeedResponse)
async def get_feed(
    user_id: str,
    limit: int = Query(default=20, ge=1, le=100),
    feed_service: IFeedProvider = Depends(_get_feed_service),
) -> FeedResponse:
    phase, videos = await feed_service.get_feed(user_id, limit)
    return FeedResponse(user_id=user_id, phase=phase, videos=videos)


@router.get("/trending", response_model=TrendingResponse)
async def get_trending(
    limit: int = Query(default=20, ge=1, le=100),
    user_id: str | None = Query(default=None),
    trending_service: ITrendingProvider = Depends(_get_trending_service),
) -> TrendingResponse:
    videos = await trending_service.get_trending(limit, user_id)
    return TrendingResponse(videos=videos)


@router.get("/health", response_model=HealthResponse)
async def health(request: Request) -> HealthResponse:
    state = request.app.state

    pg_ok = False
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


@router.post("/admin/reload-config")
async def reload_config(request: Request) -> dict:
    try:
        request.app.state.category_mapper.reload_config()
        request.app.state.scoring_config.reload_config()

        return {
            "status": "success",
            "message": "Configuration reloaded successfully"
        }
    except Exception as e:
        return {
            "status": "error",
            "message": f"Failed to reload config: {str(e)}"
        }


@router.get("/admin/categories")
async def get_all_categories(request: Request) -> dict:
    category_mapper = request.app.state.category_mapper
    categories = list(category_mapper.get_all_categories())

    return {
        "categories": categories,
        "count": len(categories)
    }


@router.get("/admin/categories/{category}/hashtags")
async def get_hashtags_for_category(request: Request, category: str) -> dict:
    category_mapper = request.app.state.category_mapper
    hashtags = category_mapper.get_hashtags_for_category(category)

    return {
        "category": category,
        "hashtags": hashtags,
        "count": len(hashtags)
    }
