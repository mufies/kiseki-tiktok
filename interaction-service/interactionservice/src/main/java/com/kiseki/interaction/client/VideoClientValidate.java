package com.kiseki.interaction.client;

import java.util.UUID;

import org.springframework.stereotype.Component;
import org.springframework.web.reactive.function.client.WebClient;
import org.springframework.web.reactive.function.client.WebClientResponseException;

import lombok.extern.slf4j.Slf4j;

@Slf4j
@Component
public class VideoClientValidate {

  private final WebClient videoClient;

  public VideoClientValidate(WebClient.Builder videoClient) {
    String baseUrl = System.getenv().getOrDefault("VIDEO_SERVICE_URL", "http://video-service:8081");
    this.videoClient = videoClient.baseUrl(baseUrl + "/videos").build();
  }

  public boolean validateVideoExists(UUID videoId) {
    try {
      videoClient.get()
          .uri("/{id}", videoId)
          .retrieve()
          .bodyToMono(Void.class)
          .block();
      return true;
    } catch (WebClientResponseException.NotFound e) {
      return false;
    } catch (Exception e) {
      log.error("Error validating video: {}", e.getMessage()); // thêm cái này
      return false;
    }
  }
}
