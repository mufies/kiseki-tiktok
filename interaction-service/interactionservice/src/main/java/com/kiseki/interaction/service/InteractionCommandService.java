package com.kiseki.interaction.service;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.kiseki.interaction.dto.request.CommentRequest;
import com.kiseki.interaction.dto.response.BookMarkedResponse;
import com.kiseki.interaction.dto.response.CommentResponse;
import com.kiseki.interaction.dto.response.LikeResponse;
import com.kiseki.interaction.entity.Interaction;
import com.kiseki.interaction.entity.InteractionType;
import com.kiseki.interaction.service.interfaces.IInteractionEventPublisher;
import com.kiseki.interaction.service.interfaces.IInteractionRepository;
import com.kiseki.interaction.service.interfaces.IUserValidationClient;
import com.kiseki.interaction.service.interfaces.IVideoValidationClient;
import com.kiseki.interaction.validation.ValidationRule;
import com.kiseki.interaction.validation.rules.CommentLengthRule;
import com.kiseki.interaction.validation.rules.CommentValidationContext;
import com.kiseki.interaction.validation.rules.EmptyContentRule;
import com.kiseki.interaction.validation.rules.UserExistsRule;
import com.kiseki.interaction.validation.rules.VideoExistsRule;

import lombok.RequiredArgsConstructor;

/**
 * Command service for interaction operations.
 */
@Service
@RequiredArgsConstructor
public class InteractionCommandService {

    private final IInteractionRepository interactionRepository;
    private final IVideoValidationClient videoValidationClient;
    private final IUserValidationClient userValidationClient;
    private final IInteractionEventPublisher eventPublisher;
    private final InteractionValidationService validationService;

    // Validation rules
    private final UserExistsRule userExistsRule;
    private final VideoExistsRule videoExistsRule;
    private final EmptyContentRule emptyContentRule;
    private final CommentLengthRule commentLengthRule;

    /**
     * Toggles a like on a video.
     * If like exists, removes it. If not, creates it.
     *
     * @param videoId The video to like/unlike
     * @param userId The user performing the action
     * @return LikeResponse with current like status and count
     */
    @Transactional
    public LikeResponse toggleLike(UUID videoId, UUID userId) {
        validateUserAndVideo(userId, videoId);

        Optional<Interaction> existingLike = interactionRepository.findByUserIdAndVideoIdAndType(
            userId, videoId, InteractionType.LIKE);

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

            // Publish event only when liking (not unliking)
            String videoOwnerId = videoValidationClient.getVideoOwnerId(videoId);
            eventPublisher.publishLikeEvent(userId.toString(), videoOwnerId, videoId.toString());
        }

        long count = interactionRepository.countByVideoIdAndType(videoId, InteractionType.LIKE);
        return new LikeResponse(videoId, isLiked, count);
    }

    /**
     * Records a view for a video.
     * Only creates if user hasn't viewed this video before.
     *
     * @param videoId The video being viewed
     * @param userId The user viewing the video
     */
    @Transactional
    public void recordView(UUID videoId, UUID userId) {
        validateUserAndVideo(userId, videoId);

        Optional<Interaction> existingView = interactionRepository.findByUserIdAndVideoIdAndType(
            userId, videoId, InteractionType.VIEW);

        if (existingView.isEmpty()) {
            Interaction view = Interaction.builder()
                .userId(userId)
                .videoId(videoId)
                .type(InteractionType.VIEW)
                .build();
            interactionRepository.save(view);
        }
    }

    /**
     * Adds a comment to a video.
     * Validates comment content and publishes event.
     *
     * @param videoId The video to comment on
     * @param userId The user commenting
     * @param request The comment request with content
     * @return CommentResponse with created comment details
     */
    @Transactional
    public CommentResponse addComment(UUID videoId, UUID userId, CommentRequest request) {
        CommentValidationContext context = CommentValidationContext.builder()
            .userId(userId)
            .videoId(videoId)
            .content(request.getContent())
            .build();

        List<ValidationRule<CommentValidationContext>> rules = List.of(
            userExistsRule,
            videoExistsRule,
            emptyContentRule,
            commentLengthRule
        );
        validationService.validate(context, rules);

        Interaction comment = Interaction.builder()
            .userId(userId)
            .videoId(videoId)
            .type(InteractionType.COMMENT)
            .content(request.getContent().trim())
            .build();

        Interaction savedComment = interactionRepository.save(comment);

        String videoOwnerId = videoValidationClient.getVideoOwnerId(videoId);
        eventPublisher.publishCommentEvent(
            userId.toString(),
            videoOwnerId,
            videoId.toString(),
            savedComment.getId().toString()
        );

        return CommentResponse.builder()
            .id(savedComment.getId())
            .userId(savedComment.getUserId())
            .videoId(savedComment.getVideoId())
            .content(savedComment.getContent())
            .createdAt(savedComment.getCreatedAt())
            .build();
    }

    /**
     * Toggles a bookmark on a video.
     * If bookmark exists, removes it. If not, creates it.
     *
     * @param videoId The video to bookmark/unbookmark
     * @param userId The user performing the action
     * @return BookMarkedResponse with current bookmark status and count
     */
    @Transactional
    public BookMarkedResponse toggleBookmark(UUID videoId, UUID userId) {
        validateUserAndVideo(userId, videoId);

        Optional<Interaction> existingBookmark = interactionRepository.findByUserIdAndVideoIdAndType(
            userId, videoId, InteractionType.BOOKMARKED);

        boolean isBookmarked;
        if (existingBookmark.isPresent()) {
            interactionRepository.delete(existingBookmark.get());
            isBookmarked = false;
        } else {
            Interaction bookmark = Interaction.builder()
                .userId(userId)
                .videoId(videoId)
                .type(InteractionType.BOOKMARKED)
                .build();
            interactionRepository.save(bookmark);
            isBookmarked = true;

            String videoOwnerId = videoValidationClient.getVideoOwnerId(videoId);
            eventPublisher.publishBookmarkEvent(userId.toString(), videoOwnerId, videoId.toString());
        }

        long count = interactionRepository.countByVideoIdAndType(videoId, InteractionType.BOOKMARKED);
        return new BookMarkedResponse(videoId, isBookmarked, count);
    }

    /**
     * Centralized validation for user and video existence.
     *
     * @param userId The user ID to validate
     * @param videoId The video ID to validate
     * @throws IllegalArgumentException if user or video doesn't exist
     */
    private void validateUserAndVideo(UUID userId, UUID videoId) {
        if (!userValidationClient.isUserExists(userId)) {
            throw new IllegalArgumentException("Invalid user ID");
        }
        if (!videoValidationClient.validateVideoExists(videoId)) {
            throw new IllegalArgumentException("Invalid video ID");
        }
    }
}
