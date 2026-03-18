package com.kiseki.interaction.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;

/**
 * Response DTO for a video that user has liked.
 * Combines interaction metadata + video metadata.
 */
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class LikedVideoResponse {
    // Interaction metadata
    private Long interactionId;
    private LocalDateTime likedAt;

    // Video metadata
    private UUID videoId;
    private String title;
    private List<String> hashtags;
    private List<String> categories;

    // Additional context
    private boolean isAvailable;  // true if video still exists
}
