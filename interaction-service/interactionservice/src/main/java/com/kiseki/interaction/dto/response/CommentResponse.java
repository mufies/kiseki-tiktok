package com.kiseki.interaction.dto.response;

import java.time.LocalDateTime;
import java.util.UUID;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CommentResponse {
  private Long id;
  private UUID userId;
  private String username;
  private String userProfileImageUrl;
  private UUID videoId;
  private String content;
  private LocalDateTime createdAt;
}
