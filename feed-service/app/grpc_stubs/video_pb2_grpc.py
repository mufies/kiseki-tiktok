

import grpc
from app.grpc_stubs.video_pb2 import (
    Video,
    GetVideoRequest,
    GetVideoResponse,
    GetVideosRequest,
    GetVideosResponse,
)


class VideoServiceStub:
    """Async-compatible gRPC client stub for VideoService."""

    def __init__(self, channel: grpc.Channel) -> None:
        self.GetVideo = channel.unary_unary(
            "/video.VideoService/GetVideo",
            request_serializer=_serialize_get_video_request,
            response_deserializer=_deserialize_get_video_response,
        )
        self.GetVideos = channel.unary_unary(
            "/video.VideoService/GetVideos",
            request_serializer=_serialize_get_videos_request,
            response_deserializer=_deserialize_get_videos_response,
        )


# ─── Protobuf wire encoding helpers ───────────────────────────────────────────
# Manual protobuf encoding/decoding to match proto/video.proto wire format.
# Field numbers from video.proto:
#   GetVideoRequest:  video_id = 1
#   GetVideosRequest: limit = 1, offset = 2
#   Video:            video_id = 1, title = 2, categories = 3, hashtags = 4
#   GetVideoResponse: video = 1
#   GetVideosResponse: videos = 1


def _encode_varint(value: int) -> bytes:
    """Encode an unsigned integer as a varint."""
    result = []
    while value > 127:
        result.append((value & 0x7F) | 0x80)
        value >>= 7
    result.append(value)
    return bytes(result)


def _decode_varint(data: bytes, pos: int) -> tuple[int, int]:
    """Decode a varint from data at position pos. Returns (value, new_pos)."""
    result = 0
    shift = 0
    while True:
        b = data[pos]
        result |= (b & 0x7F) << shift
        pos += 1
        if not (b & 0x80):
            break
        shift += 7
    return result, pos


def _encode_string(field_num: int, value: str) -> bytes:
    """Encode a string field (wire type 2: length-delimited)."""
    if not value:
        return b""
    encoded = value.encode("utf-8")
    tag = (field_num << 3) | 2  # wire type 2 = length-delimited
    return _encode_varint(tag) + _encode_varint(len(encoded)) + encoded


def _encode_int32(field_num: int, value: int) -> bytes:
    """Encode an int32 field (wire type 0: varint)."""
    if value == 0:
        return b""
    tag = (field_num << 3) | 0  # wire type 0 = varint
    return _encode_varint(tag) + _encode_varint(value)


def _encode_repeated_string(field_num: int, values: list[str]) -> bytes:
    """Encode repeated string fields."""
    result = b""
    for v in values:
        result += _encode_string(field_num, v)
    return result


def _serialize_get_video_request(req: GetVideoRequest) -> bytes:
    """Serialize GetVideoRequest to protobuf binary format."""
    return _encode_string(1, req.video_id)


def _serialize_get_videos_request(req: GetVideosRequest) -> bytes:
    """Serialize GetVideosRequest to protobuf binary format."""
    return _encode_int32(1, req.limit) + _encode_int32(2, req.offset)


def _decode_string(data: bytes, pos: int, end: int) -> str:
    """Decode a string from data[pos:end]."""
    return data[pos:end].decode("utf-8")


def _parse_video(data: bytes, start: int, end: int) -> Video:
    """Parse a Video message from protobuf bytes."""
    video_id = ""
    title = ""
    categories = []
    hashtags = []

    pos = start
    while pos < end:
        tag, pos = _decode_varint(data, pos)
        field_num = tag >> 3
        wire_type = tag & 0x7

        if wire_type == 2:  # length-delimited
            length, pos = _decode_varint(data, pos)
            field_end = pos + length
            if field_num == 1:
                video_id = _decode_string(data, pos, field_end)
            elif field_num == 2:
                title = _decode_string(data, pos, field_end)
            elif field_num == 3:
                categories.append(_decode_string(data, pos, field_end))
            elif field_num == 4:
                hashtags.append(_decode_string(data, pos, field_end))
            pos = field_end
        elif wire_type == 0:  # varint
            _, pos = _decode_varint(data, pos)
        else:
            raise ValueError(f"Unsupported wire type {wire_type}")

    return Video(video_id=video_id, title=title, categories=categories, hashtags=hashtags)


def _deserialize_get_video_response(data: bytes) -> GetVideoResponse:
    """Deserialize GetVideoResponse from protobuf binary format."""
    video = Video()
    pos = 0

    while pos < len(data):
        tag, pos = _decode_varint(data, pos)
        field_num = tag >> 3
        wire_type = tag & 0x7

        if wire_type == 2:  # length-delimited
            length, pos = _decode_varint(data, pos)
            field_end = pos + length
            if field_num == 1:  # video field
                video = _parse_video(data, pos, field_end)
            pos = field_end
        elif wire_type == 0:  # varint
            _, pos = _decode_varint(data, pos)
        else:
            raise ValueError(f"Unsupported wire type {wire_type}")

    return GetVideoResponse(video=video)


def _deserialize_get_videos_response(data: bytes) -> GetVideosResponse:
    """Deserialize GetVideosResponse from protobuf binary format."""
    videos = []
    pos = 0

    while pos < len(data):
        tag, pos = _decode_varint(data, pos)
        field_num = tag >> 3
        wire_type = tag & 0x7

        if wire_type == 2:  # length-delimited
            length, pos = _decode_varint(data, pos)
            field_end = pos + length
            if field_num == 1:  # videos repeated field
                videos.append(_parse_video(data, pos, field_end))
            pos = field_end
        elif wire_type == 0:  # varint
            _, pos = _decode_varint(data, pos)
        else:
            raise ValueError(f"Unsupported wire type {wire_type}")

    return GetVideosResponse(videos=videos)
