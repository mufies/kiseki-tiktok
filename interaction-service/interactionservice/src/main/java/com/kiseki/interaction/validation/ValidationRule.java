package com.kiseki.interaction.validation;

/**
 * Generic validation rule interface.
 *
 * @param <T> The type of object to validate
 */
@FunctionalInterface
public interface ValidationRule<T> {

    /**
     * Validates the given input.
     *
     * @param input The input to validate
     * @return ValidationResult indicating success or failure with error message
     */
    ValidationResult validate(T input);
}
