# -*- coding: utf-8 -*-
# Hand-written stubs for event.proto

from dataclasses import dataclass, field


@dataclass
class TrendingVideo:
    video_id: str = ""
    title: str = ""
    categories: list[str] = field(default_factory=list)
    hashtags: list[str] = field(default_factory=list)
    watch_count: int = 0


@dataclass
class GetTrendingRequest:
    limit: int = 20


@dataclass
class GetTrendingResponse:
    videos: list[TrendingVideo] = field(default_factory=list)
