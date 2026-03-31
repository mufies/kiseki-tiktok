"""Protocol definitions for dependency inversion principle."""
from app.protocols.repositories import (
    VideoRepository,
    ProfileRepository,
    InteractionRepository,
)
from app.protocols.services import (
    IFeedProvider,
    ITrendingProvider,
    ICategoryMapper,
    IVideoScorer,
)

__all__ = [
    "VideoRepository",
    "ProfileRepository",
    "InteractionRepository",
    "IFeedProvider",
    "ITrendingProvider",
    "ICategoryMapper",
    "IVideoScorer",
]
