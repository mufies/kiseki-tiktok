"""Unit tests for FeedOrchestrationService."""
import pytest
from unittest.mock import AsyncMock, Mock

from app.services.feed_orchestration_service import FeedOrchestrationService
from app.models import VideoResult, VideoInteraction


class MockVideoRepository:
    """Mock implementation of VideoRepository protocol."""

    def __init__(self):
        self.videos = []

    async def get_all(self) -> list[dict]:
        return self.videos

    async def get_trending(self, limit: int) -> list[dict]:
        return self.videos[:limit]


class MockProfileRepository:
    """Mock implementation of ProfileRepository protocol."""

    def __init__(self):
        self.profile = None
        self.watched = set()

    async def get_profile(self, user_id: str) -> dict | None:
        return self.profile

    async def get_watched_video_ids(self, user_id: str) -> set[str]:
        return self.watched


class MockInteractionRepository:
    """Mock implementation of InteractionRepository protocol."""

    def __init__(self):
        self.interactions = {}

    async def get_bulk_interactions(
        self,
        video_ids: list[str],
        user_id: str | None = None
    ) -> dict[str, VideoInteraction]:
        return {
            vid: self.interactions.get(vid, VideoInteraction())
            for vid in video_ids
        }


class MockVideoScorer:
    """Mock implementation of IVideoScorer protocol."""

    def __init__(self):
        self.scores = {}
        self.threshold = 10.0

    def score_video(self, video: dict, profile: dict) -> float:
        return self.scores.get(video["video_id"], 0.0)

    @property
    def cold_start_threshold(self) -> float:
        return self.threshold


class MockColdStartService:
    """Mock implementation of ColdStartService."""

    def __init__(self):
        self.cold_start_videos = []

    async def get_cold_start_feed(
        self,
        limit: int,
        user_id: str | None = None
    ) -> list[VideoResult]:
        return self.cold_start_videos[:limit]


class TestFeedOrchestrationService:
    """Test suite for FeedOrchestrationService."""

    @pytest.fixture
    def repositories(self):
        """Create mock repositories."""
        return {
            "video": MockVideoRepository(),
            "profile": MockProfileRepository(),
            "interaction": MockInteractionRepository(),
        }

    @pytest.fixture
    def services(self):
        """Create mock services."""
        return {
            "scorer": MockVideoScorer(),
            "cold_start": MockColdStartService(),
        }

    @pytest.fixture
    def orchestration_service(self, repositories, services):
        """Create FeedOrchestrationService with mocks."""
        return FeedOrchestrationService(
            video_repo=repositories["video"],
            profile_repo=repositories["profile"],
            interaction_repo=repositories["interaction"],
            scoring_service=services["scorer"],
            cold_start_service=services["cold_start"],
        )

    @pytest.mark.asyncio
    async def test_cold_start_no_profile(
        self,
        orchestration_service,
        repositories,
        services
    ):
        """Test cold start when user has no profile."""
        # Setup: no profile
        repositories["profile"].profile = None

        # Setup: cold start videos
        services["cold_start"].cold_start_videos = [
            VideoResult(video_id="v1", title="Video 1", score=100.0),
            VideoResult(video_id="v2", title="Video 2", score=50.0),
        ]

        # Execute
        phase, videos = await orchestration_service.get_feed("user123", limit=2)

        # Assert
        assert phase == "cold_start"
        assert len(videos) == 2
        assert videos[0].video_id == "v1"

    @pytest.mark.asyncio
    async def test_cold_start_low_score(
        self,
        orchestration_service,
        repositories,
        services
    ):
        """Test cold start when personalization score is too low."""
        # Setup: user has profile
        repositories["profile"].profile = {
            "categories": {"Gaming": 1.0},
            "hashtags": {}
        }

        # Setup: videos with low scores
        repositories["video"].videos = [
            {"video_id": "v1", "title": "Video 1", "categories": ["Food"], "hashtags": []},
            {"video_id": "v2", "title": "Video 2", "categories": ["Music"], "hashtags": []},
        ]

        # All videos score 0.0 (no match)
        services["scorer"].scores = {"v1": 0.0, "v2": 0.0}
        services["scorer"].threshold = 10.0  # Total score 0.0 < 10.0

        # Setup: cold start fallback
        services["cold_start"].cold_start_videos = [
            VideoResult(video_id="v1", title="Video 1", score=100.0),
        ]

        # Execute
        phase, videos = await orchestration_service.get_feed("user123", limit=2)

        # Assert
        assert phase == "cold_start"

    @pytest.mark.asyncio
    async def test_personalized_feed_success(
        self,
        orchestration_service,
        repositories,
        services
    ):
        """Test successful personalized feed generation."""
        # Setup: user profile
        repositories["profile"].profile = {
            "categories": {"Gaming": 10.0},
            "hashtags": {"gameplay": 5.0}
        }

        # Setup: videos
        repositories["video"].videos = [
            {"video_id": "v1", "title": "Gaming Video", "categories": ["Gaming"], "hashtags": ["gameplay"]},
            {"video_id": "v2", "title": "Food Video", "categories": ["Food"], "hashtags": []},
            {"video_id": "v3", "title": "Music Video", "categories": ["Music"], "hashtags": []},
        ]

        # Setup: scores (v1 has high score)
        services["scorer"].scores = {
            "v1": 15.0,  # High match
            "v2": 0.5,
            "v3": 0.3,
        }
        services["scorer"].threshold = 10.0  # Total 15.8 > 10.0

        # Execute
        phase, videos = await orchestration_service.get_feed("user123", limit=2)

        # Assert
        assert phase == "personalized"
        assert len(videos) == 2
        assert videos[0].video_id == "v1"  # Highest score first
        assert videos[0].score == 15.0

    @pytest.mark.asyncio
    async def test_filter_watched_videos(
        self,
        orchestration_service,
        repositories,
        services
    ):
        """Test that watched videos are filtered out."""
        # Setup: user profile
        repositories["profile"].profile = {
            "categories": {"Gaming": 10.0},
            "hashtags": {}
        }

        # Setup: watched videos
        repositories["profile"].watched = {"v1", "v2"}

        # Setup: all videos
        repositories["video"].videos = [
            {"video_id": "v1", "title": "Watched 1", "categories": ["Gaming"], "hashtags": []},
            {"video_id": "v2", "title": "Watched 2", "categories": ["Gaming"], "hashtags": []},
            {"video_id": "v3", "title": "New Video", "categories": ["Gaming"], "hashtags": []},
        ]

        # Setup: scores
        services["scorer"].scores = {
            "v1": 10.0,
            "v2": 8.0,
            "v3": 6.0,
        }
        services["scorer"].threshold = 5.0

        # Execute
        phase, videos = await orchestration_service.get_feed("user123", limit=3)

        # Assert
        assert phase == "personalized"
        assert len(videos) == 1  # Only v3, v1 and v2 were watched
        assert videos[0].video_id == "v3"

    @pytest.mark.asyncio
    async def test_attach_interactions(
        self,
        orchestration_service,
        repositories,
        services
    ):
        """Test that interactions are attached to videos."""
        # Setup: profile and videos
        repositories["profile"].profile = {
            "categories": {"Gaming": 10.0},
            "hashtags": {}
        }

        repositories["video"].videos = [
            {"video_id": "v1", "title": "Video 1", "categories": ["Gaming"], "hashtags": []},
        ]

        services["scorer"].scores = {"v1": 15.0}
        services["scorer"].threshold = 10.0

        # Setup: interactions
        repositories["interaction"].interactions = {
            "v1": VideoInteraction(
                like_count=100,
                comment_count=50,
                is_liked=True
            )
        }

        # Execute
        phase, videos = await orchestration_service.get_feed("user123", limit=1)

        # Assert
        assert videos[0].interactions is not None
        assert videos[0].interactions.like_count == 100
        assert videos[0].interactions.comment_count == 50
        assert videos[0].interactions.is_liked is True

    @pytest.mark.asyncio
    async def test_limit_respected(
        self,
        orchestration_service,
        repositories,
        services
    ):
        """Test that limit parameter is respected."""
        # Setup
        repositories["profile"].profile = {
            "categories": {"Gaming": 10.0},
            "hashtags": {}
        }

        repositories["video"].videos = [
            {"video_id": f"v{i}", "title": f"Video {i}", "categories": ["Gaming"], "hashtags": []}
            for i in range(50)
        ]

        services["scorer"].scores = {f"v{i}": 10.0 - i * 0.1 for i in range(50)}
        services["scorer"].threshold = 5.0

        # Execute with limit=10
        phase, videos = await orchestration_service.get_feed("user123", limit=10)

        # Assert
        assert len(videos) == 10

    @pytest.mark.asyncio
    async def test_videos_sorted_by_score(
        self,
        orchestration_service,
        repositories,
        services
    ):
        """Test that videos are returned sorted by score descending."""
        # Setup
        repositories["profile"].profile = {
            "categories": {"Gaming": 10.0},
            "hashtags": {}
        }

        repositories["video"].videos = [
            {"video_id": "v1", "title": "Video 1", "categories": ["Gaming"], "hashtags": []},
            {"video_id": "v2", "title": "Video 2", "categories": ["Gaming"], "hashtags": []},
            {"video_id": "v3", "title": "Video 3", "categories": ["Gaming"], "hashtags": []},
        ]

        # Scores in random order
        services["scorer"].scores = {
            "v1": 5.0,
            "v2": 15.0,  # Highest
            "v3": 10.0,
        }
        services["scorer"].threshold = 5.0

        # Execute
        phase, videos = await orchestration_service.get_feed("user123", limit=3)

        # Assert: sorted by score descending
        assert videos[0].video_id == "v2"
        assert videos[0].score == 15.0
        assert videos[1].video_id == "v3"
        assert videos[1].score == 10.0
        assert videos[2].video_id == "v1"
        assert videos[2].score == 5.0
