"""Interaction repository — fetches user interaction data via gRPC from Interaction Service."""
from __future__ import annotations
from typing import Any

from app.grpc_stubs.interaction_pb2 import GetBulkInteractionsRequest
from app.grpc_stubs.interaction_pb2_grpc import InteractionServiceStub


class InteractionRepository:
    def __init__(self, stub: InteractionServiceStub) -> None:
        self._stub = stub

    async def get_bulk_interactions(
        self, video_ids: list[str], user_id: str | None = None
    ) -> dict[str, Any]:
        """
        Fetch interactions for multiple videos from Interaction Service via gRPC.
        Returns dict mapping video_id -> interaction dict.
        """
        if not video_ids:
            return {}

        try:
            # Build gRPC request
            request = GetBulkInteractionsRequest(video_ids=video_ids)
            if user_id:
                request.user_id = user_id

            # Call gRPC service
            response = await self._stub.GetBulkInteractions(request)

            # Parse response
            result = {}
            for interaction in response.interactions:
                result[interaction.video_id] = {
                    "like_count": interaction.like_count,
                    "comment_count": interaction.comment_count,
                    "bookmark_count": interaction.bookmark_count,
                    "view_count": interaction.view_count,
                    "is_liked": interaction.is_liked,
                    "is_bookmarked": interaction.is_bookmarked,
                }

            return result

        except Exception as exc:
            print(f"[InteractionRepository.get_bulk_interactions] gRPC error: {exc}")
            return {}
