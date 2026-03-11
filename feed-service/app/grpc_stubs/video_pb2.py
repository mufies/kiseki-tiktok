# -*- coding: utf-8 -*-
# Hand-written stubs equivalent to: python -m grpc_tools.protoc video.proto
# Avoids requiring protoc to be installed at dev time.
# In production, regenerate with: python -m grpc_tools.protoc -I ../proto --python_out=. --grpc_python_out=. ../proto/video.proto

import grpc
from dataclasses import dataclass, field


# ─── Request / Response dataclasses ─────────────────────────────────────────

@dataclass
class Video:
    video_id: str = ""
    title: str = ""
    categories: list[str] = field(default_factory=list)
    hashtags: list[str] = field(default_factory=list)


@dataclass
class GetVideoRequest:
    video_id: str = ""


@dataclass
class GetVideoResponse:
    video: Video = field(default_factory=Video)


@dataclass
class GetVideosRequest:
    limit: int = 50
    offset: int = 0


@dataclass
class GetVideosResponse:
    videos: list[Video] = field(default_factory=list)
