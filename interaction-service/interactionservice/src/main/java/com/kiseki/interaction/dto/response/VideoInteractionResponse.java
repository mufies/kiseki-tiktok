package com.kiseki.interaction.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class VideoInteractionResponse {
    private UUID videoId;
    private long likeCount;
    private long commentCount;
    private long bookmarkCount;
    private long viewCount;
    private boolean isLiked;
    private boolean isBookmarked;
}
