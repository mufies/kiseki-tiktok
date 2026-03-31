"""Trending service - handles trending video retrieval."""
from __future__ import annotations

from app.models import VideoResult
from app.protocols.repositories import VideoRepository, InteractionRepository


class TrendingService:
    """Fetches and formats trending videos."""

    def __init__(
        self,
        video_repo: VideoRepository,
        interaction_repo: InteractionRepository,
    ) -> None:
        self._video_repo = video_repo
        self._interaction_repo = interaction_repo

    async def get_trending(
        self,
        limit: int,
        user_id: str | None = None
    ) -> list[VideoResult]:
        """Get trending videos with interaction data."""
        trending = await self._video_repo.get_trending(limit)

        results = [
            VideoResult(
                video_id=v["video_id"],
                title=v["title"],
                score=float(v.get("watch_count", 0)),
                owner=v.get("owner"),
                description=v.get("description"),
                thumbnail_url=v.get("thumbnail_url"),
            )
            for v in trending
        ]

        # Fetch interaction data
        video_ids = [v.video_id for v in results]
        interactions = await self._interaction_repo.get_bulk_interactions(video_ids, user_id)

        # Merge interaction data
        for video in results:
            interaction = interactions.get(video.video_id, {})
            video.is_liked = interaction.get("is_liked", False)
            video.is_bookmarked = interaction.get("is_bookmarked", False)
            video.like_count = interaction.get("like_count", 0)
            video.comment_count = interaction.get("comment_count", 0)
            video.bookmark_count = interaction.get("bookmark_count", 0)
            video.view_count = interaction.get("view_count", 0)

        return results
