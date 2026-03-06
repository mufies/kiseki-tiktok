package com.kiseki.interaction.service;

import java.util.List;
import java.util.Optional;
import java.util.UUID;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.kiseki.interaction.client.VideoClientValidate;
import com.kiseki.interaction.dto.request.CommentRequest;
import com.kiseki.interaction.dto.response.BookMarkedResponse;
import com.kiseki.interaction.dto.response.CommentResponse;
import com.kiseki.interaction.dto.response.LikeResponse;
import com.kiseki.interaction.entity.Interaction;
import com.kiseki.interaction.entity.InteractionType;
import com.kiseki.interaction.repository.InteractionRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class InteractionService {

  private final InteractionRepository interactionRepository;
  private final VideoClientValidate videoClientValidate;

  @Transactional
    public LikeResponse toggleLike(UUID videoId, UUID userId) {
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
    }

    long count = interactionRepository.countByVideoIdAndType(videoId, InteractionType.LIKE);
    return new LikeResponse(videoId, isLiked, count);
  }

  @Transactional
  public void recordView(UUID videoId, UUID userId) {
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
    if (!videoClientValidate.validateVideoExists(videoId)) {
      throw new IllegalArgumentException("Invalid video ID");
    }

    if (request.getContent() == null || request.getContent().trim().isEmpty()) {
      throw new IllegalArgumentException("Comment content cannot be empty");
    }

    Interaction comment = Interaction.builder()
        .userId(userId)
        .videoId(videoId)
        .type(InteractionType.COMMENT)
        .content(request.getContent().trim())
        .build();

    Interaction savedComment = interactionRepository.save(comment);

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
        .map(comment -> CommentResponse.builder()
            .id(comment.getId())
            .userId(comment.getUserId())
            .videoId(comment.getVideoId())
            .content(comment.getContent())
            .createdAt(comment.getCreatedAt())
            .build())
        .collect(Collectors.toList());
  }

  @Transactional
  public BookMarkedResponse toggleBookMarked(UUID videoId, UUID userId) {
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
    }

    long count = interactionRepository.countByVideoIdAndType(videoId, InteractionType.BOOKMARKED);
    return new BookMarkedResponse(videoId, isLiked, count);
  }
}
