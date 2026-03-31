"""FastAPI application entry point for Feed Service."""
from __future__ import annotations

from contextlib import asynccontextmanager
from typing import AsyncGenerator

import asyncpg
import redis.asyncio as aioredis
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import settings
from app.grpc_client import GrpcClients
from app.repositories.profile_repo import ProfileRepository
from app.repositories.video_repo import VideoRepository
from app.repositories.interaction_repo import InteractionRepository
from app.routers.feed import router
from app.services import (
    CategoryMappingService,
    VideoScoringService,
    ScoringConfig,
    ColdStartService,
    TrendingService,
    FeedOrchestrationService,
)


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    grpc_clients = GrpcClients(
        video_service_addr=settings.video_service_grpc,
        event_service_addr=settings.event_service_grpc,
        interaction_service_addr=settings.interaction_service_grpc,
        user_service_addr=settings.user_service_grpc,
    )

    redis_client: aioredis.Redis = aioredis.from_url(
        settings.redis_url,
        encoding="utf-8",
        decode_responses=True,
    )

    db_pool: asyncpg.Pool = await asyncpg.create_pool(
        dsn=settings.database_url,
        min_size=2,
        max_size=10,
    )

    video_repo = VideoRepository(grpc_clients.video, grpc_clients.event,grpc_clients.user)
    profile_repo = ProfileRepository(redis_client)
    interaction_repo = InteractionRepository(grpc_clients.interaction)

    category_mapper = CategoryMappingService()
    scoring_config = ScoringConfig()

    scoring_service = VideoScoringService(
        category_mapper=category_mapper,
        scoring_config=scoring_config
    )

    cold_start_service = ColdStartService(
        video_repo=video_repo,
        interaction_repo=interaction_repo,
    )

    trending_service = TrendingService(
        video_repo=video_repo,
        interaction_repo=interaction_repo,
    )

    feed_orchestration = FeedOrchestrationService(
        video_repo=video_repo,
        profile_repo=profile_repo,
        interaction_repo=interaction_repo,
        scoring_service=scoring_service,
        cold_start_service=cold_start_service,
    )

    app.state.grpc_clients = grpc_clients
    app.state.redis_client = redis_client
    app.state.db_pool = db_pool
    app.state.feed_service = feed_orchestration
    app.state.trending_service = trending_service
    app.state.category_mapper = category_mapper
    app.state.scoring_config = scoring_config

    yield

    await grpc_clients.close()
    await redis_client.aclose()
    await db_pool.close()


app = FastAPI(
    title="TikTok Feed Service",
    description="Personalized video recommendation with SOLID architecture",
    version="3.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router)
