"""
Shared gRPC channel management.
Channels are long-lived and thread/async-safe — create once at startup.
"""
from __future__ import annotations

import grpc
from app.grpc_stubs.video_pb2_grpc import VideoServiceStub
from app.grpc_stubs.event_pb2_grpc import EventServiceStub


class GrpcClients:
    def __init__(self, video_service_addr: str, event_service_addr: str) -> None:
        # Insecure channels for internal service-to-service communication
        self._video_channel = grpc.aio.insecure_channel(video_service_addr)
        self._event_channel = grpc.aio.insecure_channel(event_service_addr)

        self.video = VideoServiceStub(self._video_channel)
        self.event = EventServiceStub(self._event_channel)

    async def close(self) -> None:
        await self._video_channel.close()
        await self._event_channel.close()
