package com.kiseki.userservice.grpc;

import com.kiseki.user.grpc.GetUserRequest;
import com.kiseki.user.grpc.GetUserResponse;
import com.kiseki.user.grpc.User;
import com.kiseki.user.grpc.UserServiceGrpc;
import com.kiseki.user.grpc.VerifyUserRequest;
import com.kiseki.user.grpc.VerifyUserResponse;
import com.kiseki.user.grpc.UserFollowStatusRequest;
import com.kiseki.user.grpc.UserFollowStatusResponse;
import com.kiseki.userservice.repository.FollowRepository;
import com.kiseki.userservice.repository.UserRepository;
import com.kiseki.userservice.service.UserService;
import io.grpc.stub.StreamObserver;
import lombok.RequiredArgsConstructor;
import net.devh.boot.grpc.server.service.GrpcService;

@GrpcService
@RequiredArgsConstructor
public class UserGrpcService extends UserServiceGrpc.UserServiceImplBase {

  private final UserRepository userRepository;
  private final UserService userService;
  private final FollowRepository followRepository;

  @Override
  public void verifyUser(VerifyUserRequest request, StreamObserver<VerifyUserResponse> responseObserver) {
    // TODO: Add rate limiting to prevent user enumeration attacks
    // Example: Use Redis or in-memory cache to track requests per IP/service

    // Input validation
    if (request.getUserId() == null || request.getUserId().trim().isEmpty()) {
      responseObserver.onError(
          io.grpc.Status.INVALID_ARGUMENT
              .withDescription("UserId is required")
              .asRuntimeException());
      return;
    }

    boolean exists = userRepository.existsById(request.getUserId());

    VerifyUserResponse response = VerifyUserResponse.newBuilder()
        .setExists(exists)
        .setUserId(request.getUserId())
        .build();

    responseObserver.onNext(response);
    responseObserver.onCompleted();
  }

  @Override
  public void getUser(GetUserRequest request, StreamObserver<GetUserResponse> responseObserver) {
    try {
      // TODO: Add authorization check - ensure requester has permission to view this
      // user
      // TODO: Add rate limiting to prevent abuse

      // Input validation
      if (request.getUserId() == null || request.getUserId().trim().isEmpty()) {
        responseObserver.onError(
            io.grpc.Status.INVALID_ARGUMENT
                .withDescription("UserId is required")
                .asRuntimeException());
        return;
      }

      com.kiseki.userservice.entity.User user = userRepository.findById(request.getUserId())
          .orElseThrow(() -> new RuntimeException("User not found"));

      User grpcUser = User.newBuilder()
          .setUserId(user.getId())
          .setUsername(user.getUsername() != null ? user.getUsername() : "")
          .setEmail("") // REMOVED: Don't expose email without authorization
          .setDisplayName(user.getUsername() != null ? user.getUsername() : "")
          .setBio(user.getBio() != null ? user.getBio() : "")
          .setProfileImageUrl(user.getAvatarUrl() != null ? user.getAvatarUrl() : "")
          .setFollowersCount(0)
          .setFollowingCount(0)
          .setCreatedAt(user.getCreatedAt() != null ? user.getCreatedAt().toString() : "")
          .build();

      GetUserResponse response = GetUserResponse.newBuilder()
          .setUser(grpcUser)
          .build();

      responseObserver.onNext(response);
      responseObserver.onCompleted();
    } catch (RuntimeException e) {
      responseObserver.onError(
          io.grpc.Status.NOT_FOUND
              .withDescription("User not found")
              .asRuntimeException());
    } catch (Exception e) {
      responseObserver.onError(
          io.grpc.Status.INTERNAL
              .withDescription("Internal server error")
              .asRuntimeException());
    }
  }

  @Override
  public void checkFollowStatus(UserFollowStatusRequest request,
      StreamObserver<UserFollowStatusResponse> responseObserver) {
    // Validate user_id
    if (request.getUserId() == null || request.getUserId().trim().isEmpty()) {
      responseObserver.onError(
          io.grpc.Status.INVALID_ARGUMENT
              .withDescription("UserId is required")
              .asRuntimeException());
      return;
    }

    // Validate user_id_check
    if (request.getUserIdCheck() == null || request.getUserIdCheck().trim().isEmpty()) {
      responseObserver.onError(
          io.grpc.Status.INVALID_ARGUMENT
              .withDescription("UserIdCheck is required")
              .asRuntimeException());
      return;
    }

    // Check if user_id follows user_id_check
    boolean isFollowed = followRepository.existsByFollowerIdAndFollowingId(
        request.getUserId(), // follower ID
        request.getUserIdCheck() // following ID
    );

    UserFollowStatusResponse response = UserFollowStatusResponse.newBuilder()
        .setFollowed(isFollowed)
        .build();

    responseObserver.onNext(response);
    responseObserver.onCompleted();
  }
}
