package com.kiseki.interaction.service.adapters;

import java.util.UUID;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.client.VideoClientValidate;
import com.kiseki.interaction.grpc.VideoGrpcClient;
import com.kiseki.interaction.service.interfaces.IVideoValidationClient;

import lombok.RequiredArgsConstructor;

/**
 * Adapter that bridges existing video clients to IVideoValidationClient interface.
 */
@Component
@RequiredArgsConstructor
public class VideoValidationClientAdapter implements IVideoValidationClient {

    private final VideoClientValidate videoClientValidate;
    private final VideoGrpcClient videoGrpcClient;

    @Override
    public boolean validateVideoExists(UUID videoId) {
        return videoClientValidate.validateVideoExists(videoId);
    }

    @Override
    public String getVideoOwnerId(UUID videoId) {
        return videoGrpcClient.getVideoOwnerId(videoId.toString());
    }
}
