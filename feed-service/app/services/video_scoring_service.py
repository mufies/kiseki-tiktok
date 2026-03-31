"""Video scoring service - handles recommendation scoring algorithm."""
from __future__ import annotations
import json
from pathlib import Path
from typing import Dict

from app.protocols.services import ICategoryMapper


class ScoringConfig:
    """Configuration for scoring weights."""

    def __init__(self, config_path: str | None = None) -> None:
        if config_path is None:
            config_dir = Path(__file__).parent.parent / "config"
            config_path = str(config_dir / "scoring_weights.json")

        config = self._load_config(config_path)

        self.category_weight = config["category_weight"]
        self.hashtag_weight = config["hashtag_weight"]
        self.cold_start_threshold = config["cold_start_threshold"]
        self._config_path = config_path

        total_weight = self.category_weight + self.hashtag_weight
        if abs(total_weight - 1.0) > 0.001:
            raise ValueError(
                f"Scoring weights must sum to 1.0, got {total_weight}. "
                f"category_weight={self.category_weight}, hashtag_weight={self.hashtag_weight}"
            )

    def _load_config(self, config_path: str) -> Dict:
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                config = json.load(f)

            required_keys = {"category_weight", "hashtag_weight", "cold_start_threshold"}
            missing_keys = required_keys - set(config.keys())
            if missing_keys:
                raise ValueError(f"Missing required config keys: {missing_keys}")

            return config
        except FileNotFoundError:
            raise FileNotFoundError(
                f"Scoring weights config not found: {config_path}"
            )
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON in scoring weights: {e}")

    def reload_config(self) -> None:
        config = self._load_config(self._config_path)
        self.category_weight = config["category_weight"]
        self.hashtag_weight = config["hashtag_weight"]
        self.cold_start_threshold = config["cold_start_threshold"]


class VideoScoringService:
    """Responsible for scoring videos against user profiles."""

    def __init__(
        self,
        category_mapper: ICategoryMapper,
        scoring_config: ScoringConfig | None = None
    ) -> None:
        self._category_mapper = category_mapper
        self._config = scoring_config or ScoringConfig()

    def score_video(self, video: dict, profile: dict) -> float:
        """Calculate personalization score for a video."""
        categories_profile: dict[str, float] = profile.get("categories", {})
        hashtags_profile: dict[str, float] = profile.get("hashtags", {})

        video_categories = video.get("categories") or []
        if not video_categories:
            video_categories = self._category_mapper.get_categories(
                video.get("hashtags", [])
            )

        category_score = sum(
            categories_profile.get(cat, 0.0)
            for cat in video_categories
        )

        hashtag_score = sum(
            hashtags_profile.get(tag.lower().lstrip('#'), 0.0)
            for tag in (video.get("hashtags") or [])
        )

        final_score = (
            category_score * self._config.category_weight +
            hashtag_score * self._config.hashtag_weight
        )

        return final_score

    def score_videos_bulk(
        self,
        videos: list[dict],
        profile: dict
    ) -> list[tuple[dict, float]]:
        return [
            (video, self.score_video(video, profile))
            for video in videos
        ]

    @property
    def cold_start_threshold(self) -> float:
        return self._config.cold_start_threshold
