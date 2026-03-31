package com.kiseki.interaction.service.interfaces;

import java.util.UUID;

/**
 * Interface for user validation operations.
 */
public interface IUserValidationClient {

    boolean isUserExists(UUID userId);
}
