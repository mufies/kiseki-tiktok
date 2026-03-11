package com.kiseki.interaction.grpc;

import java.util.UUID;

import org.springframework.stereotype.Component;

import com.kiseki.user.grpc.UserServiceGrpc;
import com.kiseki.user.grpc.VerifyUserRequest;
import com.kiseki.user.grpc.VerifyUserResponse;

import io.grpc.StatusRuntimeException;
import net.devh.boot.grpc.client.inject.GrpcClient;

Slf4j
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
}
