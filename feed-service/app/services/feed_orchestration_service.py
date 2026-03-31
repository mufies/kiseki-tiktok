"""Feed orchestration service - main feed logic coordinator."""
from __future__ import annotations

from app.models import VideoResult
from app.protocols.repositories import (
    VideoRepository,
    ProfileRepository,
    InteractionRepository,
)
from app.protocols.services import IVideoScorer
from app.services.cold_start_service import ColdStartService


class FeedOrchestrationService:
    """Orchestrates personalized feed generation."""

    def __init__(
        self,
        video_repo: VideoRepository,
        profile_repo: ProfileRepository,
        interaction_repo: InteractionRepository,
        scoring_service: IVideoScorer,
        cold_start_service: ColdStartService,
    ) -> None:
        self._video_repo = video_repo
        self._profile_repo = profile_repo
        self._interaction_repo = interaction_repo
        self._scoring_service = scoring_service
        self._cold_start_service = cold_start_service

    async def get_feed(
        self,
        user_id: str,
        limit: int = 20
    ) -> tuple[str, list[VideoResult]]:
        """Get personalized feed for user."""
        profile = await self._profile_repo.get_profile(user_id)
        watched = await self._profile_repo.get_watched_video_ids(user_id)

        if profile is None:
            videos = await self._cold_start_service.get_cold_start_feed(limit, user_id)
            return "cold_start", videos

        all_videos = await self._video_repo.get_all()
        scores: list[VideoResult] = []

        for video in all_videos:
            if video["video_id"] in watched:
                continue

            score = self._scoring_service.score_video(video, profile)
            scores.append(VideoResult(
                video_id=video["video_id"],
                title=video["title"],
                score=round(score, 4),
                owner=video.get("owner"),
                description=video.get("description"),
                thumbnail_url=video.get("thumbnail_url"),
            ))

        total_score = sum(r.score for r in scores)
        if total_score < self._scoring_service.cold_start_threshold:
            videos = await self._cold_start_service.get_cold_start_feed(limit, user_id)
            return "cold_start", videos

        scores.sort(key=lambda r: r.score, reverse=True)
        top_videos = scores[:limit]

        # Fetch interaction data for top videos
        video_ids = [v.video_id for v in top_videos]
        interactions = await self._interaction_repo.get_bulk_interactions(video_ids, user_id)

        # Merge interaction data into videos
        for video in top_videos:
            interaction = interactions.get(video.video_id, {})
            video.is_liked = interaction.get("is_liked", False)
            video.is_bookmarked = interaction.get("is_bookmarked", False)
            video.like_count = interaction.get("like_count", 0)
            video.comment_count = interaction.get("comment_count", 0)
            video.bookmark_count = interaction.get("bookmark_count", 0)
            video.view_count = interaction.get("view_count", 0)

        return "personalized", top_videos
