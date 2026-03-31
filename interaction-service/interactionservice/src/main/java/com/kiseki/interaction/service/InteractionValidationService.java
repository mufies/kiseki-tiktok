package com.kiseki.interaction.service;

import java.util.List;

import org.springframework.stereotype.Service;

import com.kiseki.interaction.validation.ValidationResult;
import com.kiseki.interaction.validation.ValidationRule;

/**
 * Centralized validation service for all interaction operations.
 */
@Service
public class InteractionValidationService {

    /**
     * Validates input against a list of validation rules.
     *
     * @param input The input to validate
     * @param rules List of validation rules to apply
     * @param <T> The type of input being validated
     * @throws IllegalArgumentException if any validation rule fails
     */
    public <T> void validate(T input, List<ValidationRule<T>> rules) {
        for (ValidationRule<T> rule : rules) {
            ValidationResult result = rule.validate(input);
            if (result.isFailed()) {
                throw new IllegalArgumentException(result.getErrorMessage());
            }
        }
    }

    /**
     * Validates input against a list of validation rules and returns the result.
     *
     * @param input The input to validate
     * @param rules List of validation rules to apply
     * @param <T> The type of input being validated
     * @return ValidationResult - first failure or success if all pass
     */
    public <T> ValidationResult validateAndReturn(T input, List<ValidationRule<T>> rules) {
        for (ValidationRule<T> rule : rules) {
            ValidationResult result = rule.validate(input);
            if (result.isFailed()) {
                return result;
            }
        }
        return ValidationResult.success();
    }
}
