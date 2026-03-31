package com.kiseki.userservice.dto.response;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class UserResponse {
  private String id;
  private String email;
  private String username;
  private String avatarUrl;
  private int followerCount;
  private int followingCount;
  private LocalDateTime createdAt;
  private String bio;
}
