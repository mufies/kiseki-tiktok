package com.kiseki.interaction.client;

import java.util.List;
import java.util.Map;
import java.util.UUID;

/**
 * Interface for fetching video metadata.
 * Follows Dependency Inversion Principle - depend on abstraction, not concrete implementation.
 *
 * Can be implemented by:
 * - VideoGrpcClient (production)
 * - MockVideoClient (testing)
 * - CachedVideoClient (performance optimization)
 */
public interface VideoMetadataClient {

    /**
     * Get video metadata by ID.
     * Returns null if video doesn't exist.
     */
    VideoMetadata getVideoById(UUID videoId);

    /**
     * Bulk fetch video metadata for multiple IDs.
     * More efficient than calling getVideoById() in a loop.
     *
     * @param videoIds List of video IDs to fetch
     * @return Map of videoId -> VideoMetadata (missing videos excluded)
     */
    Map<UUID, VideoMetadata> getBulkVideos(List<UUID> videoIds);

    /**
     * Simple DTO for video metadata.
     */
    record VideoMetadata(
        UUID videoId,
        String title,
        List<String> hashtags,
        List<String> categories
    ) {}
}
