"""Cold start service - handles new user recommendations."""
from __future__ import annotations

from app.models import VideoResult
from app.protocols.repositories import VideoRepository, InteractionRepository


class ColdStartService:
    """Provides recommendations to new users using trending videos."""

    def __init__(
        self,
        video_repo: VideoRepository,
        interaction_repo: InteractionRepository,
    ) -> None:
        self._video_repo = video_repo
        self._interaction_repo = interaction_repo

    async def get_cold_start_feed(
        self,
        limit: int,
        user_id: str | None = None
    ) -> list[VideoResult]:
        """Get cold start feed using trending videos."""
        trending = await self._video_repo.get_trending(limit)

        if len(trending) < limit:
            results = await self._backfill_with_all_videos(trending, limit)
        else:
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

    async def _backfill_with_all_videos(
        self,
        trending: list[dict],
        limit: int
    ) -> list[VideoResult]:
        all_videos = await self._video_repo.get_all()

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

        trending_ids = {v["video_id"] for v in trending}
        remaining = [
            v for v in all_videos
            if v["video_id"] not in trending_ids
        ]

        results.extend([
            VideoResult(
                video_id=v["video_id"],
                title=v["title"],
                score=0.0,
                owner=v.get("owner"),
                description=v.get("description"),
                thumbnail_url=v.get("thumbnail_url"),
            )
            for v in remaining[:limit - len(trending)]
        ])

        return results
