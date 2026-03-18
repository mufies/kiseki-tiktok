"""Feed scoring service with auto-category detection from hashtags."""
from __future__ import annotations
from app.models import VideoResult
from app.repositories.profile_repo import ProfileRepository
from app.repositories.video_repo import VideoRepository
from app.repositories.interaction_repo import InteractionRepository

COLD_START_THRESHOLD = 10.0

# ─── Hashtag to Category Mapping ─────────────────────────────────────────────
HASHTAG_CATEGORY_MAP = {
    # Gaming
    "gaming": "Gaming",
    "gameplay": "Gaming", 
    "gamer": "Gaming",
    "esports": "Gaming",
    "streamer": "Gaming",
    "valorant": "Gaming",
    "lol": "Gaming",
    "minecraft": "Gaming",
    "fortnite": "Gaming",
    "fps": "Gaming",
    "moba": "Gaming",
    
    # Food & Cooking
    "cooking": "Food",
    "recipe": "Food",
    "food": "Food",
    "foodie": "Food",
    "chef": "Food",
    "baking": "Food",
    "kitchen": "Food",
    "homemade": "Food",
    "yummy": "Food",
    "delicious": "Food",
    
    # Fitness & Health
    "fitness": "Health",
    "workout": "Health",
    "gym": "Health",
    "health": "Health",
    "yoga": "Health",
    "exercise": "Health",
    "bodybuilding": "Health",
    "wellness": "Health",
    "nutrition": "Health",
    
    # Music
    "music": "Music",
    "song": "Music",
    "singing": "Music",
    "cover": "Music",
    "musician": "Music",
    "guitar": "Music",
    "piano": "Music",
    "hiphop": "Music",
    "rock": "Music",
    "pop": "Music",
    
    # Education
    "education": "Education",
    "learning": "Education",
    "tutorial": "Education",
    "howto": "Education",
    "science": "Education",
    "math": "Education",
    "history": "Education",
    "study": "Education",
    
    # Entertainment
    "comedy": "Entertainment",
    "funny": "Entertainment",
    "meme": "Entertainment",
    "prank": "Entertainment",
    "vlog": "Entertainment",
    "entertainment": "Entertainment",
    
    # Beauty & Fashion
    "beauty": "Beauty",
    "makeup": "Beauty",
    "fashion": "Beauty",
    "style": "Beauty",
    "skincare": "Beauty",
    "ootd": "Beauty",
    
    # Travel
    "travel": "Travel",
    "vacation": "Travel",
    "adventure": "Travel",
    "explore": "Travel",
    "wanderlust": "Travel",
    
    # Technology
    "tech": "Technology",
    "technology": "Technology",
    "coding": "Technology",
    "programming": "Technology",
    "ai": "Technology",
    "software": "Technology",
    "gadget": "Technology",
}


class FeedService:
    def __init__(
        self,
        video_repo: VideoRepository,
        profile_repo: ProfileRepository,
        interaction_repo: InteractionRepository,
    ) -> None:
        self._video_repo = video_repo
        self._profile_repo = profile_repo
        self._interaction_repo = interaction_repo
    
    # ─── Auto-detect categories from hashtags ─────────────────────────────────
    @staticmethod
    def extract_categories_from_hashtags(hashtags: list[str]) -> list[str]:
        """
        Auto-detect categories based on hashtags.
        Returns unique list of categories.
        """
        if not hashtags:
            return []
        
        categories = set()
        for tag in hashtags:
            # Normalize hashtag (lowercase, remove #)
            normalized_tag = tag.lower().strip().lstrip('#')
            
            # Look up category
            if category := HASHTAG_CATEGORY_MAP.get(normalized_tag):
                categories.add(category)
        
        return list(categories)
    
    # ─── Score a single video against a profile ───────────────────────────────
    @staticmethod
    def _score_video(video: dict, profile: dict) -> float:
        """
        Score video using both hashtags and auto-detected categories.
        Auto-detects categories from hashtags if not present.
        """
        categories_profile: dict[str, float] = profile.get("categories", {})
        hashtags_profile: dict[str, float] = profile.get("hashtags", {})
        
        # Get or auto-detect categories
        video_categories = video.get("categories") or []
        if not video_categories:
            # Auto-detect from hashtags
            video_categories = FeedService.extract_categories_from_hashtags(
                video.get("hashtags", [])
            )
        
        # Calculate scores
        category_score = sum(
            categories_profile.get(cat, 0.0)
            for cat in video_categories
        )
        
        hashtag_score = sum(
            hashtags_profile.get(tag.lower().lstrip('#'), 0.0)
            for tag in (video.get("hashtags") or [])
        )
        
        # Weight: categories 30%, hashtags 70%
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
            return await self._cold_start(limit, user_id)

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
            return await self._cold_start(limit, user_id)

        scores.sort(key=lambda r: r.score, reverse=True)
        top_videos = scores[:limit]

        # Fetch interactions for top videos
        video_ids = [v.video_id for v in top_videos]
        interactions = await self._interaction_repo.get_bulk_interactions(video_ids, user_id)

        # Attach interactions to videos
        for video in top_videos:
            video.interactions = interactions.get(video.video_id)

        return "personalized", top_videos
    
    # ─── Cold start: trending videos ─────────────────────────────────────────
    async def _cold_start(self, limit: int, user_id: str | None = None) -> tuple[str, list[VideoResult]]:
        trending = await self._video_repo.get_trending(limit)

        # If not enough trending videos, fallback to all videos
        if len(trending) < limit:
            all_videos = await self._video_repo.get_all()

            # Use trending first, then add remaining videos
            trending_ids = {v["video_id"] for v in trending}
            remaining = [v for v in all_videos if v["video_id"] not in trending_ids]

            # Combine: trending (with watch_count) + remaining (score = 0)
            results = [
                VideoResult(
                    video_id=v["video_id"],
                    title=v["title"],
                    score=float(v.get("watch_count", 0)),
                )
                for v in trending
            ]

            results.extend([
                VideoResult(
                    video_id=v["video_id"],
                    title=v["title"],
                    score=0.0,
                )
                for v in remaining[:limit - len(trending)]
            ])
        else:
            # Enough trending videos
            results = [
                VideoResult(
                    video_id=v["video_id"],
                    title=v["title"],
                    score=float(v.get("watch_count", 0)),
                )
                for v in trending
            ]

        # Fetch interactions for videos
        video_ids = [v.video_id for v in results]
        interactions = await self._interaction_repo.get_bulk_interactions(video_ids, user_id)

        # Attach interactions to videos
        for video in results:
            video.interactions = interactions.get(video.video_id)

        return "cold_start", results
    
  # ─── Trending endpoint ───────────────────────────────────────────────────
    async def get_trending(self, limit: int = 20, user_id: str | None = None) -> list[VideoResult]:
        trending = await self._video_repo.get_trending(limit)
        results = [
            VideoResult(
                video_id=v["video_id"],
                title=v["title"],
                score=float(v.get("watch_count", 0)),
            )
            for v in trending
        ]

        # Fetch interactions for trending videos
        video_ids = [v.video_id for v in results]
        interactions = await self._interaction_repo.get_bulk_interactions(video_ids, user_id)

        # Attach interactions to videos
        for video in results:
            video.interactions = interactions.get(video.video_id)

        return results
    
    # ─── Helper: Get video with auto-detected categories ─────────────────────
    async def get_video_with_categories(self, video_id: str) -> dict:
        """
        Fetch video and auto-detect categories if missing.
        Useful for debugging or API responses.
        """
        video = await self._video_repo.get_by_id(video_id)
        if not video:
            return None
        
        if not video.get("categories"):
            video["categories"] = self.extract_categories_from_hashtags(
                video.get("hashtags", [])
            )
        
        return video
