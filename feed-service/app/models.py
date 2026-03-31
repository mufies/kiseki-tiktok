from datetime import datetime
from pydantic import BaseModel


class VideoOwner(BaseModel):
    user_id: str
    username: str
    display_name: str | None = None
    profile_image_url: str | None = None
    followers_count: int = 0
    following_count: int = 0
    is_verified: bool = False

class VideoResult(BaseModel):
    video_id: str
    title: str
    score: float
    owner: VideoOwner | None = None
    description: str | None = None
    thumbnail_url: str | None = None
    is_liked: bool = False
    is_bookmarked: bool = False
    like_count: int = 0
    comment_count: int = 0
    bookmark_count: int = 0
    view_count: int = 0


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
