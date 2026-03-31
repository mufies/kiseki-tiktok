"""Service protocols for interface segregation."""
from __future__ import annotations
from typing import Protocol

from app.models import VideoResult


class IFeedProvider(Protocol):
    """Interface for feed generation."""

    async def get_feed(self, user_id: str, limit: int) -> tuple[str, list[VideoResult]]:
        """
        Get personalized feed for user.
        Returns (phase, videos) where phase is "personalized" or "cold_start".
        """
        ...


class ITrendingProvider(Protocol):
    """Interface for trending video retrieval."""

    async def get_trending(self, limit: int, user_id: str | None = None) -> list[VideoResult]:
        """Get trending videos."""
        ...


class ICategoryMapper(Protocol):
    """Interface for hashtag to category mapping."""

    def get_categories(self, hashtags: list[str]) -> list[str]:
        """Extract categories from hashtags."""
        ...


class IVideoScorer(Protocol):
    """Interface for video scoring algorithm."""

    def score_video(self, video: dict, profile: dict) -> float:
        """Calculate personalization score for a video against user profile."""
        ...
