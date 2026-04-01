package com.kiseki.interaction.service;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.kiseki.interaction.grpc.UserGrpcClient;
import com.kiseki.interaction.grpc.VideoGrpcClient;
import com.kiseki.interaction.client.VideoClientValidate;
import com.kiseki.interaction.client.VideoMetadataClient;
import com.kiseki.interaction.dto.request.CommentRequest;
import com.kiseki.interaction.dto.response.BookMarkedResponse;
import com.kiseki.interaction.dto.response.CommentResponse;
import com.kiseki.interaction.dto.response.LikeResponse;
import com.kiseki.interaction.dto.response.LikedVideoResponse;
import com.kiseki.interaction.dto.response.VideoInteractionResponse;
import com.kiseki.interaction.entity.Interaction;
import com.kiseki.interaction.entity.InteractionType;
import com.kiseki.interaction.kafka.KafkaProducerService;
import com.kiseki.interaction.repository.InteractionRepository;

import lombok.RequiredArgsConstructor;

/**
 * @deprecated Use InteractionCommandService for write operations
 */
@Service
@RequiredArgsConstructor
public class InteractionService {

  private final InteractionRepository interactionRepository;
  private final VideoClientValidate videoClientValidate;
  private final UserGrpcClient userGrpcClient;
  private final VideoGrpcClient videoGrpcClient;
  private final VideoMetadataClient videoMetadataClient;

  private final InteractionCommandService commandService;

  /**
   * @deprecated Use InteractionCommandService.toggleLike() instead
   */
  @Transactional
  public LikeResponse toggleLike(UUID videoId, UUID userId) {
    return commandService.toggleLike(videoId, userId);
  }

  /**
   * @deprecated Use InteractionCommandService.recordView() instead
   */
  @Transactional
  public void recordView(UUID videoId, UUID userId) {
    commandService.recordView(videoId, userId);
  }

  /**
   * @deprecated Use InteractionCommandService.addComment() instead
   */
  @Transactional
  public CommentResponse addComment(UUID videoId, UUID userId, CommentRequest request) {
    return commandService.addComment(videoId, userId, request);
  }

  @Transactional(readOnly = true)
  public LikeResponse getLikesCount(UUID videoId) {
    if (!videoClientValidate.validateVideoExists(videoId)) {
      throw new IllegalArgumentException("Invalid video ID");
    }
    long count = interactionRepository.countByVideoIdAndType(videoId, InteractionType.LIKE);
    return new LikeResponse(videoId, false, count);
  }

  @Transactional(readOnly = true)
  public List<CommentResponse> getComments(UUID videoId) {
    if (!videoClientValidate.validateVideoExists(videoId)) {
      throw new IllegalArgumentException("Invalid video ID");
    }
    List<Interaction> comments = interactionRepository.findByVideoIdAndTypeOrderByCreatedAtDesc(videoId,
        InteractionType.COMMENT);

    return comments.stream()
        .map(comment -> {
          var userInfo = userGrpcClient.getUserById(comment.getUserId());

          return CommentResponse.builder()
              .id(comment.getId())
              .userId(comment.getUserId())
              .username(userInfo != null ? userInfo.getUsername() : "Unknown User")
              .userProfileImageUrl(userInfo != null ? userInfo.getProfileImageUrl() : null)
              .videoId(comment.getVideoId())
              .content(comment.getContent())
              .createdAt(comment.getCreatedAt())
              .build();
        })
        .collect(Collectors.toList());
  }

  /**
   * @deprecated Use InteractionCommandService.toggleBookmark() instead
   */
  @Transactional
  public BookMarkedResponse toggleBookMarked(UUID videoId, UUID userId) {
    return commandService.toggleBookmark(videoId, userId);
  }

  /**
   * Get interaction data for a single video.
   * This method is optimized for single video queries.
   *
   * @param videoId The video ID to get interactions for
   * @param userId Optional user ID to check if user has liked/bookmarked
   * @return VideoInteractionResponse with all interaction counts and user-specific data
   */
  @Transactional(readOnly = true)
  public VideoInteractionResponse getVideoInteraction(UUID videoId, UUID userId) {
    // Get all interaction counts for this video
    long likeCount = interactionRepository.countByVideoIdAndType(videoId, InteractionType.LIKE);
    long commentCount = interactionRepository.countByVideoIdAndType(videoId, InteractionType.COMMENT);
    long bookmarkCount = interactionRepository.countByVideoIdAndType(videoId, InteractionType.BOOKMARKED);
    long viewCount = interactionRepository.countByVideoIdAndType(videoId, InteractionType.VIEW);

    // Check if current user has liked/bookmarked this video
    boolean isLiked = false;
    boolean isBookmarked = false;

    if (userId != null) {
      isLiked = interactionRepository.existsByUserIdAndVideoIdAndType(userId, videoId, InteractionType.LIKE);
      isBookmarked = interactionRepository.existsByUserIdAndVideoIdAndType(userId, videoId, InteractionType.BOOKMARKED);
    }

    return VideoInteractionResponse.builder()
        .videoId(videoId)
        .likeCount(likeCount)
        .commentCount(commentCount)
        .bookmarkCount(bookmarkCount)
        .viewCount(viewCount)
        .isLiked(isLiked)
        .isBookmarked(isBookmarked)
        .build();
  }

  @Transactional(readOnly = true)
  public List<VideoInteractionResponse> getBulkInteractions(List<UUID> videoIds, UUID userId) {
    if (videoIds == null || videoIds.isEmpty()) {
      return List.of();
    }

    List<Object[]> counts = interactionRepository.countInteractionsByVideoIds(videoIds);
    Map<UUID, Map<InteractionType, Long>> countMap = new HashMap<>();

    for (Object[] row : counts) {
      UUID videoId = (UUID) row[0];
      InteractionType type = (InteractionType) row[1];
      Long count = (Long) row[2];

      countMap.computeIfAbsent(videoId, k -> new HashMap<>()).put(type, count);
    }

    Map<UUID, Boolean> userLikes = new HashMap<>();
    Map<UUID, Boolean> userBookmarks = new HashMap<>();

    if (userId != null) {
      List<Interaction> userInteractions = interactionRepository.findByUserIdAndVideoIdInAndType(
          userId, videoIds, InteractionType.LIKE);
      userInteractions.forEach(i -> userLikes.put(i.getVideoId(), true));

      List<Interaction> userBookmarkInteractions = interactionRepository.findByUserIdAndVideoIdInAndType(
          userId, videoIds, InteractionType.BOOKMARKED);
      userBookmarkInteractions.forEach(i -> userBookmarks.put(i.getVideoId(), true));
    }

    return videoIds.stream().map(videoId -> {
      Map<InteractionType, Long> videoCounts = countMap.getOrDefault(videoId, new HashMap<>());

      return VideoInteractionResponse.builder()
          .videoId(videoId)
          .likeCount(videoCounts.getOrDefault(InteractionType.LIKE, 0L))
          .commentCount(videoCounts.getOrDefault(InteractionType.COMMENT, 0L))
          .bookmarkCount(videoCounts.getOrDefault(InteractionType.BOOKMARKED, 0L))
          .viewCount(videoCounts.getOrDefault(InteractionType.VIEW, 0L))
          .isLiked(userLikes.getOrDefault(videoId, false))
          .isBookmarked(userBookmarks.getOrDefault(videoId, false))
          .build();
    }).collect(Collectors.toList());
  }

  /**
   * Get all videos that a user has liked.
   *
   * @param userId User ID to fetch liked videos for
   * @return List of videos with interaction metadata, ordered by most recent like
   */
  @Transactional(readOnly = true)
  public List<LikedVideoResponse> getUserLikedVideos(UUID userId) {
    List<Interaction> likes = interactionRepository
        .findByUserIdAndTypeOrderByCreatedAtDesc(userId, InteractionType.LIKE);

    if (likes.isEmpty()) {
      return List.of();
    }

    List<UUID> videoIds = likes.stream()
        .map(Interaction::getVideoId)
        .collect(Collectors.toList());

    Map<UUID, VideoMetadataClient.VideoMetadata> videoMetadataMap = videoMetadataClient.getBulkVideos(videoIds);

    return likes.stream()
        .map(like -> {
          VideoMetadataClient.VideoMetadata videoMeta = videoMetadataMap.get(like.getVideoId());

          if (videoMeta != null) {
            return LikedVideoResponse.builder()
                .interactionId(like.getId())
                .likedAt(like.getCreatedAt())
                .videoId(videoMeta.videoId())
                .title(videoMeta.title())
                .hashtags(videoMeta.hashtags())
                .categories(videoMeta.categories())
                .isAvailable(true)
                .build();
          } else {
            return LikedVideoResponse.builder()
                .interactionId(like.getId())
                .likedAt(like.getCreatedAt())
                .videoId(like.getVideoId())
                .title("Video Unavailable")
                .hashtags(List.of())
                .categories(List.of())
                .isAvailable(false)
                .build();
          }
        })
        .collect(Collectors.toList());
  }

}
