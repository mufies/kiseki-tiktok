from datetime import datetime
from pydantic import BaseModel


class VideoResult(BaseModel):
    video_id: str
    title: str
    score: float


class FeedResponse(BaseModel):
    user_id: str
    phase: str  # "personalized" | "cold_start"
    videos: list[VideoResult]


class TrendingResponse(BaseModel):
    videos: list[VideoResult]


class HealthResponse(BaseModel):
    service: str = "feed-service"
    version: str = "1.0"
    status: str
    postgres: str
    redis: str
