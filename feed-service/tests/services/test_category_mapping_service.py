"""Unit tests for CategoryMappingService."""
import pytest
import json
import tempfile
from pathlib import Path

from app.services.category_mapping_service import CategoryMappingService


class TestCategoryMappingService:
    """Test suite for category mapping service."""

    @pytest.fixture
    def sample_mapping(self):
        """Sample category mapping for testing."""
        return {
            "gaming": "Gaming",
            "gameplay": "Gaming",
            "food": "Food",
            "cooking": "Food",
            "music": "Music",
        }

    @pytest.fixture
    def mapping_service(self, sample_mapping):
        """Create a CategoryMappingService with temp config."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(sample_mapping, f)
            temp_path = f.name

        service = CategoryMappingService(mapping_config_path=temp_path)
        yield service

        # Cleanup
        Path(temp_path).unlink()

    def test_get_categories_single_hashtag(self, mapping_service):
        """Test extracting category from single hashtag."""
        categories = mapping_service.get_categories(["gaming"])
        assert categories == ["Gaming"]

    def test_get_categories_multiple_hashtags_same_category(self, mapping_service):
        """Test multiple hashtags mapping to same category."""
        categories = mapping_service.get_categories(["gaming", "gameplay"])
        assert categories == ["Gaming"]  # Deduplicated

    def test_get_categories_multiple_hashtags_different_categories(self, mapping_service):
        """Test multiple hashtags mapping to different categories."""
        categories = mapping_service.get_categories(["gaming", "food", "music"])
        assert set(categories) == {"Gaming", "Food", "Music"}

    def test_get_categories_with_hash_prefix(self, mapping_service):
        """Test hashtags with # prefix are handled correctly."""
        categories = mapping_service.get_categories(["#gaming", "#food"])
        assert set(categories) == {"Gaming", "Food"}

    def test_get_categories_case_insensitive(self, mapping_service):
        """Test hashtags are case-insensitive."""
        categories = mapping_service.get_categories(["GAMING", "GaMiNg", "gaming"])
        assert categories == ["Gaming"]

    def test_get_categories_unknown_hashtag(self, mapping_service):
        """Test unknown hashtags are ignored."""
        categories = mapping_service.get_categories(["unknown", "notfound"])
        assert categories == []

    def test_get_categories_mixed_known_unknown(self, mapping_service):
        """Test mix of known and unknown hashtags."""
        categories = mapping_service.get_categories(["gaming", "unknown", "food"])
        assert set(categories) == {"Gaming", "Food"}

    def test_get_categories_empty_list(self, mapping_service):
        """Test empty hashtag list returns empty categories."""
        categories = mapping_service.get_categories([])
        assert categories == []

    def test_get_all_categories(self, mapping_service):
        """Test getting all available categories."""
        all_categories = mapping_service.get_all_categories()
        assert all_categories == {"Gaming", "Food", "Music"}

    def test_get_hashtags_for_category(self, mapping_service):
        """Test reverse lookup: category -> hashtags."""
        hashtags = mapping_service.get_hashtags_for_category("Gaming")
        assert set(hashtags) == {"gaming", "gameplay"}

    def test_reload_config(self, sample_mapping):
        """Test config can be reloaded without restart."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            json.dump(sample_mapping, f)
            temp_path = f.name

        service = CategoryMappingService(mapping_config_path=temp_path)

        # Initial state
        assert service.get_categories(["gaming"]) == ["Gaming"]
        assert service.get_categories(["tech"]) == []

        # Update config file
        updated_mapping = {**sample_mapping, "tech": "Technology"}
        with open(temp_path, 'w') as f:
            json.dump(updated_mapping, f)

        # Reload
        service.reload_config()

        # New mapping should be active
        assert service.get_categories(["tech"]) == ["Technology"]

        # Cleanup
        Path(temp_path).unlink()

    def test_invalid_config_file_not_found(self):
        """Test error handling for missing config file."""
        with pytest.raises(FileNotFoundError):
            CategoryMappingService(mapping_config_path="/nonexistent/path.json")

    def test_invalid_config_malformed_json(self):
        """Test error handling for malformed JSON."""
        with tempfile.NamedTemporaryFile(mode='w', suffix='.json', delete=False) as f:
            f.write("{ invalid json }")
            temp_path = f.name

        with pytest.raises(ValueError, match="Invalid JSON"):
            CategoryMappingService(mapping_config_path=temp_path)

        Path(temp_path).unlink()

    def test_default_config_path(self):
        """Test service uses default config path when none provided."""
        # This test assumes the default config exists
        # Skip if file not found
        try:
            service = CategoryMappingService()
            # Should load without error
            assert service.mapping is not None
        except FileNotFoundError:
            pytest.skip("Default config file not found")
