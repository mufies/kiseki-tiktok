package com.kiseki.interaction.grpc;

import java.util.UUID;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.dto.UserInfoDTO;
import com.kiseki.user.grpc.GetUserRequest;
import com.kiseki.user.grpc.GetUserResponse;
import com.kiseki.user.grpc.User;
import com.kiseki.user.grpc.UserServiceGrpc;
import com.kiseki.user.grpc.VerifyUserRequest;
import com.kiseki.user.grpc.VerifyUserResponse;
import com.kiseki.user.grpc.UserServiceGrpc.UserServiceBlockingStub;
import io.grpc.StatusRuntimeException;
import lombok.extern.slf4j.Slf4j;
import net.devh.boot.grpc.client.inject.GrpcClient;

@Slf4j
@Component
public class UserGrpcClient {

  @GrpcClient("user-service")
  private UserServiceGrpc.UserServiceBlockingStub userServiceStub;

  public boolean isUserExists(UUID userId) {
    try {
      VerifyUserRequest request = VerifyUserRequest.newBuilder()
          .setUserId(userId.toString())
          .build();

      VerifyUserResponse response = userServiceStub.verifyUser(request);

      log.debug("User verification for ID {}: exists={}", userId, response.getExists());
      return response.getExists();
    } catch (StatusRuntimeException e) {
      log.error("gRPC error while verifying user {}: {}", userId, e.getStatus());
      return false;
    } catch (Exception e) {
      log.error("Error while verifying user {}: {}", userId, e.getMessage());
      return false;
    }
  }

  public UserInfoDTO getUserById(UUID userId) {
    try {
      GetUserRequest request = GetUserRequest.newBuilder()
          .setUserId(userId.toString())
          .build();

      GetUserResponse response = userServiceStub.getUser(request);

      if (response != null && response.hasUser()) {
        User userProto = response.getUser();
        return UserInfoDTO.builder()
            .userId(UUID.fromString(userProto.getUserId()))
            .username(userProto.getUsername())
            .profileImageUrl(userProto.getProfileImageUrl())
            .build();
      } else {
        log.warn("User with ID {} does not exist", userId);
        return null;
      }
    } catch (StatusRuntimeException e) {
      log.error("gRPC error while fetching user {}: {}", userId, e.getStatus());
      return null;
    } catch (Exception e) {
      log.error("Error while fetching user {}: {}", userId, e.getMessage());
      return null;
    }
  }
}
