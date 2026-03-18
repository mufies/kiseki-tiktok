package com.kiseki.interaction.client;

import com.kiseki.interaction.grpc.VideoGrpcClient;
import com.kiseki.user.grpc.User;
import com.kiseki.video.grpc.GetVideoRequest;
import com.kiseki.video.grpc.GetVideoResponse;
import com.kiseki.video.grpc.Video;
import com.kiseki.video.grpc.VideoServiceGrpc;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;
import org.springframework.stereotype.Component;

import java.util.*;
import java.util.stream.Collectors;

/**
 * gRPC implementation of VideoMetadataClient.
 * Adapter pattern: wraps gRPC client to implement our interface.
 */
@Slf4j
@Component
@RequiredArgsConstructor
public class VideoGrpcMetadataClient implements VideoMetadataClient {

    @GrpcClient("video-service")
    private VideoServiceGrpc.VideoServiceBlockingStub videoServiceStub;

    @Override
    public VideoMetadata getVideoById(UUID videoId) {
        try {
            GetVideoRequest request = GetVideoRequest.newBuilder()
                .setVideoId(videoId.toString())
                .build();

            GetVideoResponse response = videoServiceStub.getVideo(request);

            if (response.hasVideo()) {
                Video video = response.getVideo();
                return new VideoMetadata(
                    UUID.fromString(video.getVideoId()),
                    video.getTitle(),
                    new ArrayList<>(video.getHashtagsList()),
                    new ArrayList<>(video.getCategoriesList())
                );
            }

            return null;
        } catch (Exception e) {
            log.error("Failed to fetch video metadata for videoId: {}", videoId, e);
            return null;
        }
    }

    @Override
    public Map<UUID, VideoMetadata> getBulkVideos(List<UUID> videoIds) {
        // Note: Ideally Video Service should have GetBulkVideos gRPC endpoint
        // For now, fallback to individual calls (can be optimized later)

        Map<UUID, VideoMetadata> result = new HashMap<>();

        for (UUID videoId : videoIds) {
            VideoMetadata metadata = getVideoById(videoId);
            if (metadata != null) {
                result.put(videoId, metadata);
            }
        }

        log.debug("Fetched {} out of {} videos", result.size(), videoIds.size());
        return result;
    }
}
