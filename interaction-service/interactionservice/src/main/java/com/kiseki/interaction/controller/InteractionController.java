package com.kiseki.interaction.controller;

import com.kiseki.interaction.dto.request.CommentRequest;
import com.kiseki.interaction.dto.response.BookMarkedResponse;
import com.kiseki.interaction.dto.response.CommentResponse;
import com.kiseki.interaction.dto.response.LikeResponse;
import com.kiseki.interaction.dto.response.LikedVideoResponse;
import com.kiseki.interaction.dto.response.VideoInteractionResponse;
import com.kiseki.interaction.service.InteractionCommandService;
import com.kiseki.interaction.service.InteractionService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.UUID;

@RestController
@RequestMapping("/interactions/videos")
@RequiredArgsConstructor
public class InteractionController {

  private final InteractionCommandService interactionCommandService;
  private final InteractionService interactionService;

  @PostMapping("/{videoId}/like")
  public ResponseEntity<LikeResponse> toggleLike(
      @PathVariable UUID videoId,
      @RequestHeader("X-User-Id") UUID userId) {
    LikeResponse response = interactionCommandService.toggleLike(videoId, userId);
    return ResponseEntity.ok(response);
  }

  @PostMapping("/{videoId}/bookmarked")
  public ResponseEntity<BookMarkedResponse> toggleBookMarked(
      @PathVariable UUID videoId,
      @RequestHeader("X-User-Id") UUID userId) {
    BookMarkedResponse response = interactionCommandService.toggleBookmark(videoId, userId);
    return ResponseEntity.ok(response);
  }

  @PostMapping("/{videoId}/view")
  public ResponseEntity<Void> recordView(
      @PathVariable UUID videoId,
      @RequestHeader("X-User-Id") UUID userId) {
    interactionCommandService.recordView(videoId, userId);
    return ResponseEntity.status(HttpStatus.CREATED).build();
  }

  @PostMapping("/{videoId}/comment")
  public ResponseEntity<CommentResponse> addComment(
      @PathVariable UUID videoId,
      @RequestHeader("X-User-Id") UUID userId,
      @RequestBody CommentRequest request) {
    CommentResponse response = interactionCommandService.addComment(videoId, userId, request);
    return ResponseEntity.status(HttpStatus.CREATED).body(response);
  }

  @GetMapping("/{videoId}/likes")
  public ResponseEntity<LikeResponse> getLikesCount(@PathVariable UUID videoId) {
    LikeResponse response = interactionService.getLikesCount(videoId);
    return ResponseEntity.ok(response);
  }

  @GetMapping("/{videoId}/comments")
  public ResponseEntity<List<CommentResponse>> getComments(@PathVariable UUID videoId) {
    List<CommentResponse> response = interactionService.getComments(videoId);
    return ResponseEntity.ok(response);
  }

  @GetMapping("/bulk")
  public ResponseEntity<List<VideoInteractionResponse>> getBulkInteractions(
      @RequestParam List<UUID> videoIds,
      @RequestHeader(value = "X-User-Id", required = false) UUID userId) {
    List<VideoInteractionResponse> response = interactionService.getBulkInteractions(videoIds, userId);
    return ResponseEntity.ok(response);
  }

  @GetMapping("/users/{userId}/liked-videos")
  public ResponseEntity<List<LikedVideoResponse>> getUserLikedVideos(@PathVariable UUID userId) {
    List<LikedVideoResponse> response = interactionService.getUserLikedVideos(userId);
    return ResponseEntity.ok(response);
  }
}
