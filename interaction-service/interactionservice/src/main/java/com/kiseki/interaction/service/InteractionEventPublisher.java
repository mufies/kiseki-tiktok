package com.kiseki.interaction.service;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.service.interfaces.IInteractionEventPublisher;
import com.kiseki.interaction.kafka.KafkaProducerService;

import lombok.RequiredArgsConstructor;

/**
 * Service responsible for publishing interaction events to Kafka.
 */
@Component
@RequiredArgsConstructor
public class InteractionEventPublisher implements IInteractionEventPublisher {

    private final KafkaProducerService kafkaProducerService;

    @Override
    public void publishLikeEvent(String userId, String videoOwnerId, String videoId) {
        if (videoOwnerId != null && !videoOwnerId.equals(userId)) {
            kafkaProducerService.sendLikeEvent(userId, videoOwnerId, videoId);
        }
    }

    @Override
    public void publishCommentEvent(String userId, String videoOwnerId, String videoId, String commentId) {
        if (videoOwnerId != null && !videoOwnerId.equals(userId)) {
            kafkaProducerService.sendCommentEvent(userId, videoOwnerId, videoId, commentId);
        }
    }

    @Override
    public void publishBookmarkEvent(String userId, String videoOwnerId, String videoId) {
        if (videoOwnerId != null && !videoOwnerId.equals(userId)) {
            kafkaProducerService.sendBookmarkEvent(userId, videoOwnerId, videoId);
        }
    }
}
