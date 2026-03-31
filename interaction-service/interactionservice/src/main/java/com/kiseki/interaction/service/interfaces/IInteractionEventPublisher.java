package com.kiseki.interaction.service.interfaces;

/**
 * Interface for publishing interaction events.
 */
public interface IInteractionEventPublisher {

    void publishLikeEvent(String userId, String videoOwnerId, String videoId);

    void publishCommentEvent(String userId, String videoOwnerId, String videoId, String commentId);

    void publishBookmarkEvent(String userId, String videoOwnerId, String videoId);
}
