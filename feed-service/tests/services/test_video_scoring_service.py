"""Unit tests for VideoScoringService."""
import pytest
import json
import tempfile
from pathlib import Path
from unittest.mock import Mock

from app.services.video_scoring_service import VideoScoringService, ScoringConfig


class MockCategoryMapper:
    """Mock implementation of ICategoryMapper protocol."""

    def get_categories(self, hashtags: list[str]) -> list[str]:
        """Mock category mapping."""
        mapping = {
            "gaming": "Gaming",
            "food": "Food",
            "music": "Music",
        }
        categories = []
        for tag in hashtags:
            normalized = tag.lower().lstrip('#')
            if cat := mapping.get(normalized):
                if cat not in categories:
                    categories.append(cat)
        return categories


class TestScoringConfig:
    """Test suite for ScoringConfig."""

    def test_load_valid_config(self):
        """Test loading valid scoring config."""
        config_data = {
            "category_weight": 0.3,
            "hashtag_weight": 0.7,
            "cold_start_threshold": 10.0
        }

        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(config_data, f)
            temp_path = f.name

        config = ScoringConfig(config_path=temp_path)

        assert config.category_weight == 0.3
        assert config.hashtag_weight == 0.7
        assert config.cold_start_threshold == 10.0

        Path(temp_path).unlink()

    def test_config_weights_must_sum_to_one(self):
        """Test validation that weights sum to 1.0."""
        config_data = {
            "category_weight": 0.5,
            "hashtag_weight": 0.3,  # Sum = 0.8, invalid
            "cold_start_threshold": 10.0
        }

        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(config_data, f)
            temp_path = f.name

        with pytest.raises(ValueError, match="must sum to 1.0"):
            ScoringConfig(config_path=temp_path)

        Path(temp_path).unlink()

    def test_config_missing_required_keys(self):
        """Test error when required keys are missing."""
        config_data = {
            "category_weight": 0.3,
            # Missing hashtag_weight and cold_start_threshold
        }

        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(config_data, f)
            temp_path = f.name

        with pytest.raises(ValueError, match="Missing required config keys"):
            ScoringConfig(config_path=temp_path)

        Path(temp_path).unlink()

    def test_reload_config(self):
        """Test config can be reloaded."""
        initial_config = {
            "category_weight": 0.3,
            "hashtag_weight": 0.7,
            "cold_start_threshold": 10.0
        }

        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(initial_config, f)
            temp_path = f.name

        config = ScoringConfig(config_path=temp_path)
        assert config.category_weight == 0.3

        # Update config file
        updated_config = {
            "category_weight": 0.4,
            "hashtag_weight": 0.6,
            "cold_start_threshold": 15.0
        }
        with open(temp_path, 'w') as f:
            json.dump(updated_config, f)

        config.reload_config()

        assert config.category_weight == 0.4
        assert config.hashtag_weight == 0.6
        assert config.cold_start_threshold == 15.0

        Path(temp_path).unlink()


class TestVideoScoringService:
    """Test suite for VideoScoringService."""

    @pytest.fixture
    def scoring_config(self):
        """Create a test scoring config."""
        config_data = {
            "category_weight": 0.3,
            "hashtag_weight": 0.7,
            "cold_start_threshold": 10.0
        }

        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(config_data, f)
            temp_path = f.name

        config = ScoringConfig(config_path=temp_path)
        yield config

        Path(temp_path).unlink()

    @pytest.fixture
    def scoring_service(self, scoring_config):
        """Create a VideoScoringService for testing."""
        category_mapper = MockCategoryMapper()
        return VideoScoringService(
            category_mapper=category_mapper,
            scoring_config=scoring_config
        )

    def test_score_video_with_category_match(self, scoring_service):
        """Test scoring when video category matches user profile."""
        video = {
            "video_id": "v1",
            "categories": ["Gaming"],
            "hashtags": []
        }
        profile = {
            "categories": {"Gaming": 5.0},
            "hashtags": {}
        }

        score = scoring_service.score_video(video, profile)

        # 5.0 * 0.3 (category_weight) + 0 * 0.7 (hashtag_weight)
        assert score == pytest.approx(1.5)

    def test_score_video_with_hashtag_match(self, scoring_service):
        """Test scoring when video hashtag matches user profile."""
        video = {
            "video_id": "v1",
            "categories": [],
            "hashtags": ["gaming", "esports"]
        }
        profile = {
            "categories": {},
            "hashtags": {"gaming": 3.0, "esports": 2.0}
        }

        score = scoring_service.score_video(video, profile)

        # 0 * 0.3 + (3.0 + 2.0) * 0.7
        assert score == pytest.approx(3.5)

    def test_score_video_with_both_matches(self, scoring_service):
        """Test scoring with both category and hashtag matches."""
        video = {
            "video_id": "v1",
            "categories": ["Gaming"],
            "hashtags": ["gaming"]
        }
        profile = {
            "categories": {"Gaming": 10.0},
            "hashtags": {"gaming": 5.0}
        }

        score = scoring_service.score_video(video, profile)

        # 10.0 * 0.3 + 5.0 * 0.7
        assert score == pytest.approx(6.5)

    def test_score_video_auto_detect_categories(self, scoring_service):
        """Test auto-detection of categories from hashtags."""
        video = {
            "video_id": "v1",
            # No categories provided
            "hashtags": ["gaming", "food"]
        }
        profile = {
            "categories": {"Gaming": 5.0, "Food": 3.0},
            "hashtags": {}
        }

        score = scoring_service.score_video(video, profile)

        # Auto-detected categories: Gaming, Food
        # (5.0 + 3.0) * 0.3 + 0 * 0.7
        assert score == pytest.approx(2.4)

    def test_score_video_no_match(self, scoring_service):
        """Test scoring when no categories or hashtags match."""
        video = {
            "video_id": "v1",
            "categories": ["Technology"],
            "hashtags": ["tech"]
        }
        profile = {
            "categories": {"Gaming": 5.0},
            "hashtags": {"gaming": 3.0}
        }

        score = scoring_service.score_video(video, profile)

        assert score == pytest.approx(0.0)

    def test_score_video_hashtag_case_insensitive(self, scoring_service):
        """Test hashtag matching is case-insensitive."""
        video = {
            "video_id": "v1",
            "categories": [],
            "hashtags": ["GAMING", "#Gaming"]
        }
        profile = {
            "categories": {},
            "hashtags": {"gaming": 5.0}
        }

        score = scoring_service.score_video(video, profile)

        # Both hashtags normalize to "gaming", score counted once per occurrence
        # (5.0 + 5.0) * 0.7
        assert score == pytest.approx(7.0)

    def test_score_videos_bulk(self, scoring_service):
        """Test bulk scoring multiple videos."""
        videos = [
            {"video_id": "v1", "categories": ["Gaming"], "hashtags": []},
            {"video_id": "v2", "categories": ["Food"], "hashtags": []},
            {"video_id": "v3", "categories": ["Music"], "hashtags": []},
        ]
        profile = {
            "categories": {"Gaming": 10.0, "Food": 5.0},
            "hashtags": {}
        }

        results = scoring_service.score_videos_bulk(videos, profile)

        assert len(results) == 3
        assert results[0][1] == pytest.approx(3.0)  # Gaming: 10.0 * 0.3
        assert results[1][1] == pytest.approx(1.5)  # Food: 5.0 * 0.3
        assert results[2][1] == pytest.approx(0.0)  # Music: no match

    def test_cold_start_threshold_property(self, scoring_service):
        """Test accessing cold start threshold."""
        assert scoring_service.cold_start_threshold == 10.0
