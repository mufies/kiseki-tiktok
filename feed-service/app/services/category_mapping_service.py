"""Category mapping service - handles hashtag to category conversion."""
from __future__ import annotations
import json
import os
from pathlib import Path
from typing import Dict


class CategoryMappingService:
    """Responsible for mapping hashtags to categories."""

    def __init__(self, mapping_config_path: str | None = None) -> None:
        if mapping_config_path is None:
            config_dir = Path(__file__).parent.parent / "config"
            mapping_config_path = str(config_dir / "category_mapping.json")

        self.mapping = self._load_mapping(mapping_config_path)
        self._config_path = mapping_config_path

    def _load_mapping(self, config_path: str) -> Dict[str, str]:
        try:
            with open(config_path, 'r', encoding='utf-8') as f:
                mapping = json.load(f)

            if not isinstance(mapping, dict):
                raise ValueError("Category mapping must be a dictionary")

            return mapping
        except FileNotFoundError:
            raise FileNotFoundError(
                f"Category mapping config not found: {config_path}"
            )
        except json.JSONDecodeError as e:
            raise ValueError(f"Invalid JSON in category mapping: {e}")

    def get_categories(self, hashtags: list[str]) -> list[str]:
        """Extract categories from hashtags."""
        if not hashtags:
            return []

        categories = set()
        for tag in hashtags:
            normalized_tag = tag.lower().strip().lstrip('#')

            if category := self.mapping.get(normalized_tag):
                categories.add(category)

        return list(categories)

    def reload_config(self) -> None:
        self.mapping = self._load_mapping(self._config_path)

    def get_all_categories(self) -> set[str]:
        return set(self.mapping.values())

    def get_hashtags_for_category(self, category: str) -> list[str]:
        return [
            hashtag
            for hashtag, cat in self.mapping.items()
            if cat == category
        ]
