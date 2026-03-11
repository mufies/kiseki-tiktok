package com.kiseki.interaction.kafka;

import com.kiseki.interaction.dto.NotificationEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
@Slf4j
public class KafkaProducerService {

    private final KafkaTemplate<String, NotificationEvent> kafkaTemplate;

    public void sendLikeEvent(String fromUserId, String toUserId, String videoId) {
        NotificationEvent event = NotificationEvent.builder()
                .type("LIKE")
                .fromUserId(fromUserId)
                .toUserId(toUserId)
                .videoId(videoId)
                .build();

        kafkaTemplate.send("interaction.liked", event);
        log.info("Sent like notification event: {} liked video {} owned by {}", fromUserId, videoId, toUserId);
    }

    public void sendCommentEvent(String fromUserId, String toUserId, String videoId, String commentId) {
        NotificationEvent event = NotificationEvent.builder()
                .type("COMMENT")
                .fromUserId(fromUserId)
                .toUserId(toUserId)
                .videoId(videoId)
                .commentId(commentId)
                .build();

        kafkaTemplate.send("interaction.commented", event);
        log.info("Sent comment notification event: {} commented on video {} owned by {}", fromUserId, videoId, toUserId);
    }

    public void sendBookmarkEvent(String fromUserId, String toUserId, String videoId) {
        NotificationEvent event = NotificationEvent.builder()
                .type("BOOKMARK")
                .fromUserId(fromUserId)
                .toUserId(toUserId)
                .videoId(videoId)
                .build();

        kafkaTemplate.send("interaction.bookmarked", event);
        log.info("Sent bookmark notification event: {} bookmarked video {} owned by {}", fromUserId, videoId, toUserId);
    }
}
