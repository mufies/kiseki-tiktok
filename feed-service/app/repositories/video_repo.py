"""Video repository — fetches video metadata via gRPC from Video Service."""
from __future__ import annotations

from app.grpc_stubs.video_pb2 import GetVideosRequest, GetVideoRequest
from app.grpc_stubs.video_pb2_grpc import VideoServiceStub
from app.grpc_stubs.event_pb2 import GetTrendingRequest
from app.grpc_stubs.event_pb2_grpc import EventServiceStub
from app.grpc_stubs.user_pb2 import GetUserRequest                                               
from app.grpc_stubs.user_pb2_grpc import UserServiceStub   
from app.models import VideoOwner


class VideoRepository:
    def __init__(
        self,
        video_stub: VideoServiceStub,
        event_stub: EventServiceStub,
        user_stub: UserServiceStub,
    ) -> None:
        self._user = user_stub
        self._video = video_stub
        self._event = event_stub

    async def get_all(self) -> list[dict]:
        """Fetch all videos from Video Service via gRPC."""
        try:
            response = await self._video.GetVideos(GetVideosRequest(limit=500, offset=0))
            results = []
            for v in response.videos:
                owner = await self._convert_owner(v.owner) if v.HasField("owner") else None
                results.append({
                    "video_id": v.video_id,
                    "title": v.title,
                    "categories": list(v.categories),
                    "hashtags": list(v.hashtags),
                    "description": v.description,
                    "thumbnail_url": v.thumbnail_url,
                    "owner": owner,
                })
            return results
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
            results = []
            for v in response.videos:
                owner = await self._convert_owner(v.owner) if v.HasField("owner") else None
                results.append({
                    "video_id": v.video_id,
                    "title": v.title,
                    "categories": list(v.categories),
                    "hashtags": list(v.hashtags),
                    "description": v.description if hasattr(v, "description") else "",
                    "thumbnail_url": v.thumbnail_url if hasattr(v, "thumbnail_url") else "",
                    "watch_count": v.watch_count,
                    "owner": owner,
                })
            return results
        except Exception as exc:
            print(f"[VideoRepository.get_trending] gRPC error: {exc}")
            return []

    async def _convert_owner(self, owner) -> VideoOwner | None:
        """Convert protobuf VideoOwner to model VideoOwner."""
        if not owner:
            return None

        # Gọi User Service để lấy thông tin đầy đủ
        try:
            response = await self._user.GetUser(GetUserRequest(user_id=owner.user_id))
            user = response.user
            return VideoOwner(
                user_id=user.user_id,
                username=user.username,
                display_name=user.display_name if user.display_name else None,
                profile_image_url=user.profile_image_url if user.profile_image_url else None,
                followers_count=user.followers_count,
                following_count=user.following_count,
                is_verified=getattr(owner, 'is_verified', False),
            )
        except Exception as exc:
            # Nếu gọi User Service lỗi, fallback dùng data có sẵn từ owner
            print(f"[VideoRepository._convert_owner] Failed to fetch user {owner.user_id}: {exc}")
            return VideoOwner(
                user_id=owner.user_id,
                username=owner.username if hasattr(owner, 'username') else "",
                display_name=owner.display_name if owner.display_name else None,
                profile_image_url=owner.profile_image_url if owner.profile_image_url else None,
                followers_count=owner.followers_count if hasattr(owner, 'followers_count') else 0,
                following_count=owner.following_count if hasattr(owner, 'following_count') else 0,
                is_verified=getattr(owner, 'is_verified', False),
            )
