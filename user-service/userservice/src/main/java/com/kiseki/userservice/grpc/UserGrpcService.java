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
            com.kiseki.userservice.entity.User user = userRepository.findById(request.getUserId())
                    .orElseThrow(() -> new RuntimeException("User not found"));

            User grpcUser = User.newBuilder()
                    .setUserId(user.getId())
                    .setUsername(user.getUsername() != null ? user.getUsername() : "")
                    .setEmail(user.getEmail() != null ? user.getEmail() : "")
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
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }
}
