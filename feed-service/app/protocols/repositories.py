"""Repository protocols for dependency inversion."""
from __future__ import annotations
from typing import Protocol, Any


class VideoRepository(Protocol):
    """Protocol for video data access."""

    async def get_all(self) -> list[dict]:
        """Fetch all videos."""
        ...

    async def get_trending(self, limit: int) -> list[dict]:
        """Fetch trending videos."""
        ...


class ProfileRepository(Protocol):
    """Protocol for user profile data access."""

    async def get_profile(self, user_id: str) -> dict | None:
        """Get user profile with preferences."""
        ...

    async def get_watched_video_ids(self, user_id: str) -> set[str]:
        """Get list of video IDs the user has watched."""
        ...


class InteractionRepository(Protocol):
    """Protocol for video interaction data access."""

    async def get_bulk_interactions(
        self,
        video_ids: list[str],
        user_id: str | None = None
    ) -> dict[str, Any]:
        """Fetch interactions for multiple videos."""
        ...
