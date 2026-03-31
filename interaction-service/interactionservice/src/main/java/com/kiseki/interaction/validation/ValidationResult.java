package com.kiseki.interaction.validation;

import lombok.AllArgsConstructor;
import lombok.Getter;

/**
 * Represents the result of a validation operation.
 */
@Getter
@AllArgsConstructor
public class ValidationResult {
    private final boolean valid;
    private final String errorMessage;

    public static ValidationResult success() {
        return new ValidationResult(true, null);
    }

    public static ValidationResult failure(String errorMessage) {
        return new ValidationResult(false, errorMessage);
    }

    public boolean isFailed() {
        return !valid;
    }
}
