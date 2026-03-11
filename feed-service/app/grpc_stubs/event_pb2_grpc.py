# -*- coding: utf-8 -*-
# Hand-written gRPC stub for EventService (client side — Feed Service calls Event Service)
# Uses proper Protocol Buffer wire format to communicate with C# server.

import grpc
from app.grpc_stubs.event_pb2 import GetTrendingRequest, GetTrendingResponse, TrendingVideo


class EventServiceStub:
    """gRPC client stub for EventService."""

    def __init__(self, channel: grpc.Channel) -> None:
        self.GetTrendingVideos = channel.unary_unary(
            "/event.EventService/GetTrendingVideos",
            request_serializer=_serialize_trending_request,
            response_deserializer=_deserialize_trending_response,
        )


# ─── Protobuf wire encoding helpers ───────────────────────────────────────────
# Manual protobuf encoding/decoding to match proto/event.proto wire format.
# Field numbers from event.proto:
#   GetTrendingRequest: limit = 1
#   TrendingVideo:      video_id = 1, title = 2, categories = 3, hashtags = 4, watch_count = 5
#   GetTrendingResponse: videos = 1


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


def _encode_int32(field_num: int, value: int) -> bytes:
    """Encode an int32 field (wire type 0: varint)."""
    if value == 0:
        return b""
    tag = (field_num << 3) | 0  # wire type 0 = varint
    return _encode_varint(tag) + _encode_varint(value)


def _decode_string(data: bytes, pos: int, end: int) -> str:
    """Decode a string from data[pos:end]."""
    return data[pos:end].decode("utf-8")


def _serialize_trending_request(req: GetTrendingRequest) -> bytes:
    """Serialize GetTrendingRequest to protobuf binary format."""
    return _encode_int32(1, req.limit)


def _parse_trending_video(data: bytes, start: int, end: int) -> TrendingVideo:
    """Parse a TrendingVideo message from protobuf bytes."""
    video_id = ""
    title = ""
    categories = []
    hashtags = []
    watch_count = 0

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
            value, pos = _decode_varint(data, pos)
            if field_num == 5:
                watch_count = value
        else:
            raise ValueError(f"Unsupported wire type {wire_type}")

    return TrendingVideo(
        video_id=video_id,
        title=title,
        categories=categories,
        hashtags=hashtags,
        watch_count=watch_count,
    )


def _deserialize_trending_response(data: bytes) -> GetTrendingResponse:
    """Deserialize GetTrendingResponse from protobuf binary format."""
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
                videos.append(_parse_trending_video(data, pos, field_end))
            pos = field_end
        elif wire_type == 0:  # varint
            _, pos = _decode_varint(data, pos)
        else:
            raise ValueError(f"Unsupported wire type {wire_type}")

    return GetTrendingResponse(videos=videos)
