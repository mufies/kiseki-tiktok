package com.kiseki.userservice.kafka;

import com.kiseki.userservice.dto.FollowEvent;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.kafka.core.KafkaTemplate;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
@Slf4j
public class KafkaProducerService {

    private final KafkaTemplate<String, FollowEvent> kafkaTemplate;

    public void sendFollowEvent(String fromUserId, String toUserId) {
        FollowEvent event = FollowEvent.builder()
                .type("FOLLOW")
                .fromUserId(fromUserId)
                .toUserId(toUserId)
                .build();

        kafkaTemplate.send("user.followed", event);
        log.info("Sent follow notification event: {} followed {}", fromUserId, toUserId);
    }
}
