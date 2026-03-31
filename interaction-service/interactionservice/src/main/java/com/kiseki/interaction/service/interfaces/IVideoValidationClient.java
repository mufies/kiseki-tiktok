package com.kiseki.interaction.service.interfaces;

import java.util.UUID;

/**
 * Interface for video validation operations.
 */
public interface IVideoValidationClient {

    boolean validateVideoExists(UUID videoId);

    String getVideoOwnerId(UUID videoId);
}
