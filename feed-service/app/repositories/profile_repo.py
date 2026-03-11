"""Profile repository — read user profile + watch history from Redis."""
from __future__ import annotations

import json

import redis.asyncio as aioredis


class ProfileRepository:
    def __init__(self, client: aioredis.Redis) -> None:  # type: ignore[type-arg]
        self._client = client

    async def get_profile(self, user_id: str) -> dict | None:
        """Read profile:{user_id} from Redis; returns None if missing."""
        raw = await self._client.get(f"profile:{user_id}")
        if raw is None:
            return None
        try:
            return json.loads(raw)
        except json.JSONDecodeError:
            return None

    async def get_watched_video_ids(self, user_id: str) -> set[str]:
        """
        Read history:{user_id} from Redis.
        Supports both a JSON list and a Redis Set (SADD).
        Falls back to empty set if key absent.
        """
        key = f"history:{user_id}"
        key_type = await self._client.type(key)

        if key_type == b"set":
            members = await self._client.smembers(key)
            return {m.decode() if isinstance(m, bytes) else m for m in members}

        raw = await self._client.get(key)
        if raw is None:
            return set()
        try:
            data = json.loads(raw)
            return set(data) if isinstance(data, list) else set()
        except json.JSONDecodeError:
            return set()
