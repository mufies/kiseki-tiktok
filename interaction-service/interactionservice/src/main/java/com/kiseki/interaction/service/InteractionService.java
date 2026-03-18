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

@Service
@RequiredArgsConstructor
public class InteractionService {

  private final InteractionRepository interactionRepository;
  private final VideoClientValidate videoClientValidate;
  private final UserGrpcClient userGrpcClient;
  private final VideoGrpcClient videoGrpcClient;
  private final VideoMetadataClient videoMetadataClient;
  private final KafkaProducerService kafkaProducerService;

  @Transactional
  public LikeResponse toggleLike(UUID videoId, UUID userId) {
    if (!userGrpcClient.isUserExists(userId)) {
      throw new IllegalArgumentException("Invalid user ID");
    }
    if (!videoClientValidate.validateVideoExists(videoId)) {
      throw new IllegalArgumentException("Invalid video ID");
    }
    Optional<Interaction> existingLike = interactionRepository.findByUserIdAndVideoIdAndType(userId, videoId,
        InteractionType.LIKE);

    boolean isLiked;
    if (existingLike.isPresent()) {
      interactionRepository.delete(existingLike.get());
      isLiked = false;
    } else {
      Interaction like = Interaction.builder()
          .userId(userId)
          .videoId(videoId)
          .type(InteractionType.LIKE)
          .build();
      interactionRepository.save(like);
      isLiked = true;

      // Send notification event only when liking (not unliking)
      String videoOwnerId = videoGrpcClient.getVideoOwnerId(videoId.toString());
      if (videoOwnerId != null && !videoOwnerId.equals(userId.toString())) {
        // Don't notify if user likes their own video
        kafkaProducerService.sendLikeEvent(userId.toString(), videoOwnerId, videoId.toString());
      }
    }

    long count = interactionRepository.countByVideoIdAndType(videoId, InteractionType.LIKE);
    return new LikeResponse(videoId, isLiked, count);
  }

  @Transactional
  public void recordView(UUID videoId, UUID userId) {
    if (!userGrpcClient.isUserExists(userId)) {
      throw new IllegalArgumentException("Invalid user ID");
    }
    if (!videoClientValidate.validateVideoExists(videoId)) {
      throw new IllegalArgumentException("Invalid video ID");
    }
    Optional<Interaction> existingView = interactionRepository.findByUserIdAndVideoIdAndType(userId, videoId,
        InteractionType.VIEW);

    if (existingView.isEmpty()) {
      Interaction view = Interaction.builder()
          .userId(userId)
          .videoId(videoId)
          .type(InteractionType.VIEW)
          .build();
      interactionRepository.save(view);
    }
  }

  @Transactional
  public CommentResponse addComment(UUID videoId, UUID userId, CommentRequest request) {
    if (!userGrpcClient.isUserExists(userId)) {
      throw new IllegalArgumentException("Invalid user ID");
    }
    if (!videoClientValidate.validateVideoExists(videoId)) {
      throw new IllegalArgumentException("Invalid video ID");
    }

    // Enhanced validation for fraud prevention
    if (request.getContent() == null || request.getContent().trim().isEmpty()) {
      throw new IllegalArgumentException("Comment content cannot be empty");
    }

    // Prevent DoS with extremely long comments
    if (request.getContent().length() > 1000) {
      throw new IllegalArgumentException("Comment content cannot exceed 1000 characters");
    }

    // Prevent spam - check for minimum meaningful content length
    if (request.getContent().trim().length() < 1) {
      throw new IllegalArgumentException("Comment content too short");
    }

    Interaction comment = Interaction.builder()
        .userId(userId)
        .videoId(videoId)
        .type(InteractionType.COMMENT)
        .content(request.getContent().trim())
        .build();

    Interaction savedComment = interactionRepository.save(comment);

    // Send notification event
    String videoOwnerId = videoGrpcClient.getVideoOwnerId(videoId.toString());
    if (videoOwnerId != null && !videoOwnerId.equals(userId.toString())) {
      // Don't notify if user comments on their own video
      kafkaProducerService.sendCommentEvent(
          userId.toString(),
          videoOwnerId,
          videoId.toString(),
          savedComment.getId().toString());
    }

    return CommentResponse.builder()
        .id(savedComment.getId())
        .userId(savedComment.getUserId())
        .videoId(savedComment.getVideoId())
        .content(savedComment.getContent())
        .createdAt(savedComment.getCreatedAt())
        .build();
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
          // Fetch user info via gRPC
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

  @Transactional
  public BookMarkedResponse toggleBookMarked(UUID videoId, UUID userId) {
    if (!userGrpcClient.isUserExists(userId)) {
      throw new IllegalArgumentException("Invalid user ID");
    }
    if (!videoClientValidate.validateVideoExists(videoId)) {
      throw new IllegalArgumentException("Invalid video ID");
    }
    Optional<Interaction> existingLike = interactionRepository.findByUserIdAndVideoIdAndType(userId, videoId,
        InteractionType.BOOKMARKED);

    boolean isLiked;
    if (existingLike.isPresent()) {
      interactionRepository.delete(existingLike.get());
      isLiked = false;
    } else {
      Interaction like = Interaction.builder()
          .userId(userId)
          .videoId(videoId)
          .type(InteractionType.BOOKMARKED)
          .build();
      interactionRepository.save(like);
      isLiked = true;

      // Send notification event only when bookmarking (not unbookmarking)
      String videoOwnerId = videoGrpcClient.getVideoOwnerId(videoId.toString());
      if (videoOwnerId != null && !videoOwnerId.equals(userId.toString())) {
        // Don't notify if user bookmarks their own video
        kafkaProducerService.sendBookmarkEvent(userId.toString(), videoOwnerId, videoId.toString());
      }
    }

    long count = interactionRepository.countByVideoIdAndType(videoId, InteractionType.BOOKMARKED);
    return new BookMarkedResponse(videoId, isLiked, count);
  }

  @Transactional(readOnly = true)
  public List<VideoInteractionResponse> getBulkInteractions(List<UUID> videoIds, UUID userId) {
    if (videoIds == null || videoIds.isEmpty()) {
      return List.of();
    }

    // Get all interaction counts grouped by video and type
    List<Object[]> counts = interactionRepository.countInteractionsByVideoIds(videoIds);
    Map<UUID, Map<InteractionType, Long>> countMap = new HashMap<>();

    for (Object[] row : counts) {
      UUID videoId = (UUID) row[0];
      InteractionType type = (InteractionType) row[1];
      Long count = (Long) row[2];

      countMap.computeIfAbsent(videoId, k -> new HashMap<>()).put(type, count);
    }

    // Get user's interactions with these videos
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

    // Build response for each video
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
   * Follows SOLID principles:
   * - Single Responsibility: Only orchestrates data fetching, delegates to specialized components
   * - Dependency Inversion: Depends on VideoMetadataClient interface, not concrete implementation
   * - Open/Closed: Can extend with caching/filtering without modifying this method
   *
   * @param userId User ID to fetch liked videos for
   * @return List of videos with interaction metadata, ordered by most recent like
   */
  @Transactional(readOnly = true)
  public List<LikedVideoResponse> getUserLikedVideos(UUID userId) {
    // Step 1: Get all like interactions for user (ordered by most recent)
    List<Interaction> likes = interactionRepository
        .findByUserIdAndTypeOrderByCreatedAtDesc(userId, InteractionType.LIKE);

    if (likes.isEmpty()) {
      return List.of();
    }

    // Step 2: Extract video IDs
    List<UUID> videoIds = likes.stream()
        .map(Interaction::getVideoId)
        .collect(Collectors.toList());

    // Step 3: Bulk fetch video metadata (efficient - 1 call instead of N)
    Map<UUID, VideoMetadataClient.VideoMetadata> videoMetadataMap =
        videoMetadataClient.getBulkVideos(videoIds);

    // Step 4: Combine interaction data + video metadata
    return likes.stream()
        .map(like -> {
          VideoMetadataClient.VideoMetadata videoMeta = videoMetadataMap.get(like.getVideoId());

          if (videoMeta != null) {
            // Video still exists
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
            // Video was deleted or unavailable
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
