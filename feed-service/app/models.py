from datetime import datetime
from pydantic import BaseModel


class VideoInteraction(BaseModel):
    like_count: int = 0
    comment_count: int = 0
    bookmark_count: int = 0
    view_count: int = 0
    is_liked: bool = False
    is_bookmarked: bool = False


class VideoResult(BaseModel):
    video_id: str
    title: str
    score: float
    interactions: VideoInteraction | None = None


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
