package com.kiseki.userservice.grpc;

import com.kiseki.user.grpc.GetUserRequest;
import com.kiseki.user.grpc.GetUserResponse;
import com.kiseki.user.grpc.User;
import com.kiseki.user.grpc.UserServiceGrpc;
import com.kiseki.user.grpc.VerifyUserRequest;
import com.kiseki.user.grpc.VerifyUserResponse;
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

    @Override
    public void verifyUser(VerifyUserRequest request, StreamObserver<VerifyUserResponse> responseObserver) {
        // TODO: Add rate limiting to prevent user enumeration attacks
        // Example: Use Redis or in-memory cache to track requests per IP/service

        // Input validation
        if (request.getUserId() == null || request.getUserId().trim().isEmpty()) {
            responseObserver.onError(
                io.grpc.Status.INVALID_ARGUMENT
                    .withDescription("UserId is required")
                    .asRuntimeException()
            );
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
            // TODO: Add authorization check - ensure requester has permission to view this user
            // TODO: Add rate limiting to prevent abuse

            // Input validation
            if (request.getUserId() == null || request.getUserId().trim().isEmpty()) {
                responseObserver.onError(
                    io.grpc.Status.INVALID_ARGUMENT
                        .withDescription("UserId is required")
                        .asRuntimeException()
                );
                return;
            }

            com.kiseki.userservice.entity.User user = userRepository.findById(request.getUserId())
                    .orElseThrow(() -> new RuntimeException("User not found"));

            // SECURITY FIX: Don't expose email in public profile
            // Email should only be returned if the requester is the user themselves or an admin
            User grpcUser = User.newBuilder()
                    .setUserId(user.getId())
                    .setUsername(user.getUsername() != null ? user.getUsername() : "")
                    .setEmail("") // REMOVED: Don't expose email without authorization
                    .setDisplayName(user.getUsername() != null ? user.getUsername() : "")
                    .setBio("")
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
                    .asRuntimeException()
            );
        } catch (Exception e) {
            responseObserver.onError(
                io.grpc.Status.INTERNAL
                    .withDescription("Internal server error")
                    .asRuntimeException()
            );
        }
    }
}
