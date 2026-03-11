"""Feed scoring service."""
from __future__ import annotations

from app.models import VideoResult
from app.repositories.profile_repo import ProfileRepository
from app.repositories.video_repo import VideoRepository

COLD_START_THRESHOLD = 10.0


class FeedService:
    def __init__(self, video_repo: VideoRepository, profile_repo: ProfileRepository) -> None:
        self._video_repo = video_repo
        self._profile_repo = profile_repo

    # ─── Score a single video against a profile ───────────────────────────────
    @staticmethod
    def _score_video(video: dict, profile: dict) -> float:
        categories_profile: dict[str, float] = profile.get("categories", {})
        hashtags_profile: dict[str, float] = profile.get("hashtags", {})

        category_score = sum(
            categories_profile.get(cat, 0.0)
            for cat in (video.get("categories") or [])
        )
        hashtag_score = sum(
            hashtags_profile.get(tag, 0.0)
            for tag in (video.get("hashtags") or [])
        )

        return category_score * 0.3 + hashtag_score * 0.7

    # ─── Personalized feed ────────────────────────────────────────────────────
    async def get_feed(self, user_id: str, limit: int = 20) -> tuple[str, list[VideoResult]]:
        """
        Returns (phase, videos).
        phase = "personalized" | "cold_start"
        """
        profile = await self._profile_repo.get_profile(user_id)
        watched = await self._profile_repo.get_watched_video_ids(user_id)

        # ── Cold start detection ──
        if profile is None:
            return await self._cold_start(limit)

        videos = await self._video_repo.get_all()
        scores: list[VideoResult] = []

        for v in videos:
            if v["video_id"] in watched:
                continue
            score = self._score_video(v, profile)
            scores.append(VideoResult(
                video_id=v["video_id"],
                title=v["title"],
                score=round(score, 4),
            ))

        total_score = sum(r.score for r in scores)
        if total_score < COLD_START_THRESHOLD:
            return await self._cold_start(limit)

        scores.sort(key=lambda r: r.score, reverse=True)
        return "personalized", scores[:limit]

    # ─── Cold start: trending videos ─────────────────────────────────────────
    async def _cold_start(self, limit: int) -> tuple[str, list[VideoResult]]:
        trending = await self._video_repo.get_trending(limit)
        results = [
            VideoResult(
                video_id=v["video_id"],
                title=v["title"],
                score=float(v.get("watch_count", 0)),
            )
            for v in trending
        ]
        return "cold_start", results

    # ─── Trending endpoint ───────────────────────────────────────────────────
    async def get_trending(self, limit: int = 20) -> list[VideoResult]:
        trending = await self._video_repo.get_trending(limit)
        return [
            VideoResult(
                video_id=v["video_id"],
                title=v["title"],
                score=float(v.get("watch_count", 0)),
            )
            for v in trending
        ]
