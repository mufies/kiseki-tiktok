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
    this.videoClient = videoClient.baseUrl("http://localhost:8080/api/videos").build();
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
