"""Video repository — fetches video metadata via gRPC from Video Service."""
from __future__ import annotations

from app.grpc_stubs.video_pb2 import GetVideosRequest, GetVideoRequest
from app.grpc_stubs.video_pb2_grpc import VideoServiceStub
from app.grpc_stubs.event_pb2 import GetTrendingRequest
from app.grpc_stubs.event_pb2_grpc import EventServiceStub


class VideoRepository:
    def __init__(
        self,
        video_stub: VideoServiceStub,
        event_stub: EventServiceStub,
    ) -> None:
        self._video = video_stub
        self._event = event_stub

    async def get_all(self) -> list[dict]:
        """Fetch all videos from Video Service via gRPC."""
        try:
            response = await self._video.GetVideos(GetVideosRequest(limit=500, offset=0))
            return [
                {
                    "video_id": v.video_id,
                    "title": v.title,
                    "categories": list(v.categories),
                    "hashtags": list(v.hashtags),
                }
                for v in response.videos
            ]
        except Exception as exc:
            # Log and degrade gracefully — return empty list, not crash
            print(f"[VideoRepository.get_all] gRPC error: {exc}")
            return []

    async def get_trending(self, limit: int = 20) -> list[dict]:
        """
        Get trending videos from Event Service via gRPC.
        Event Service computes trending by counting watch events.
        """
        try:
            response = await self._event.GetTrendingVideos(GetTrendingRequest(limit=limit))
            return [
                {
                    "video_id": v.video_id,
                    "title": v.title,
                    "categories": list(v.categories),
                    "hashtags": list(v.hashtags),
                    "watch_count": v.watch_count,
                }
                for v in response.videos
            ]
        except Exception as exc:
            print(f"[VideoRepository.get_trending] gRPC error: {exc}")
            return []
