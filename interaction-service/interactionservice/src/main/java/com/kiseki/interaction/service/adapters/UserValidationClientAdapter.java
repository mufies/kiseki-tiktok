package com.kiseki.interaction.service.adapters;

import java.util.UUID;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.grpc.UserGrpcClient;
import com.kiseki.interaction.service.interfaces.IUserValidationClient;

import lombok.RequiredArgsConstructor;

/**
 * Adapter that bridges existing UserGrpcClient to IUserValidationClient interface.
 */
@Component
@RequiredArgsConstructor
public class UserValidationClientAdapter implements IUserValidationClient {

    private final UserGrpcClient userGrpcClient;

    @Override
    public boolean isUserExists(UUID userId) {
        return userGrpcClient.isUserExists(userId);
    }
}
