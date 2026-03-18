package com.kiseki.interaction.grpc;

import com.kiseki.interaction.dto.response.VideoInteractionResponse;
import com.kiseki.interaction.service.InteractionService;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.server.service.GrpcService;

import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

import com.kiseki.interaction.grpc.GetBulkInteractionsRequest;
import com.kiseki.interaction.grpc.GetBulkInteractionsResponse;
import com.kiseki.interaction.grpc.InteractionServiceGrpc;
import com.kiseki.interaction.grpc.VideoInteraction;

@Slf4j
@GrpcService
@RequiredArgsConstructor
public class InteractionGrpcService extends InteractionServiceGrpc.InteractionServiceImplBase {

  private final InteractionService interactionService;

  @Override
  public void getBulkInteractions(
      GetBulkInteractionsRequest request,
      StreamObserver<GetBulkInteractionsResponse> responseObserver) {

    try {
      log.info("gRPC getBulkInteractions called with {} video IDs", request.getVideoIdsCount());

      // Parse video IDs
      List<UUID> videoIds = request.getVideoIdsList().stream()
          .map(UUID::fromString)
          .collect(Collectors.toList());

      // Parse user ID (optional)
      UUID userId = null;
      if (request.hasUserId() && !request.getUserId().isEmpty()) {
        userId = UUID.fromString(request.getUserId());
      }

      // Get interactions from service
      List<VideoInteractionResponse> interactions = interactionService.getBulkInteractions(videoIds, userId);

      // Build gRPC response
      GetBulkInteractionsResponse.Builder responseBuilder = GetBulkInteractionsResponse.newBuilder();

      for (VideoInteractionResponse interaction : interactions) {
        VideoInteraction grpcInteraction = VideoInteraction.newBuilder()
            .setVideoId(interaction.getVideoId().toString())
            .setLikeCount(interaction.getLikeCount())
            .setCommentCount(interaction.getCommentCount())
            .setBookmarkCount(interaction.getBookmarkCount())
            .setViewCount(interaction.getViewCount())
            .setIsLiked(interaction.isLiked())
            .setIsBookmarked(interaction.isBookmarked())
            .build();

        responseBuilder.addInteractions(grpcInteraction);
      }

      GetBulkInteractionsResponse response = responseBuilder.build();
      responseObserver.onNext(response);
      responseObserver.onCompleted();

      log.info("Returned {} video interactions", response.getInteractionsCount());

    } catch (Exception e) {
      log.error("Error in getBulkInteractions: {}", e.getMessage(), e);
      responseObserver.onError(e);
    }
  }
}
