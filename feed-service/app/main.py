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
from app.services.feed_service import FeedService


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    # ── Startup ──────────────────────────────────────────────────────────────

    # gRPC channels to upstream services
    grpc_clients = GrpcClients(
        video_service_addr=settings.video_service_grpc,
        event_service_addr=settings.event_service_grpc,
        interaction_service_addr=settings.interaction_service_grpc,
    )

    # Redis (profile + watch-history)
    redis_client: aioredis.Redis = aioredis.from_url(
        settings.redis_url,
        encoding="utf-8",
        decode_responses=True,
    )

    # PostgreSQL pool (still needed by ProfileRepository for history fallback)
    db_pool: asyncpg.Pool = await asyncpg.create_pool(
        dsn=settings.database_url,
        min_size=2,
        max_size=10,
    )

    video_repo       = VideoRepository(grpc_clients.video, grpc_clients.event)
    profile_repo     = ProfileRepository(redis_client)
    interaction_repo = InteractionRepository(grpc_clients.interaction)
    feed_service     = FeedService(video_repo, profile_repo, interaction_repo)

    app.state.grpc_clients  = grpc_clients
    app.state.redis_client  = redis_client
    app.state.db_pool       = db_pool
    app.state.feed_service  = feed_service

    yield

    # ── Shutdown ─────────────────────────────────────────────────────────────
    await grpc_clients.close()
    await redis_client.aclose()
    await db_pool.close()


app = FastAPI(
    title="TikTok Feed Service",
    description="Personalized video recommendation — backed by gRPC service mesh",
    version="2.1",
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
