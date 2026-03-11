package com.kiseki.interaction.client;

import java.util.UUID;

import org.springframework.stereotype.Component;
import org.springframework.web.reactive.function.client.WebClient;
import org.springframework.web.reactive.function.client.WebClientResponseException;

@Component
public class UserClientValidation {

  private final WebClient userClient;

  public UserClientValidation(WebClient.Builder builder) {
    this.userClient = builder.baseUrl("http://localhost:8080/api/users").build();
  }

  public boolean isUserExists(UUID userId) {
    try {
      userClient.get()
          .uri("/verify/{id}", userId)
          .retrieve()
          .bodyToMono(Object.class) // Expect detailed response
          .block();
      return true;
    } catch (WebClientResponseException.NotFound e) {
      return false;
    } catch (Exception e) {
      return false;
    }
  }
}
