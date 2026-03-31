package com.kiseki.interaction.validation.rules;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.service.interfaces.IUserValidationClient;
import com.kiseki.interaction.validation.ValidationResult;
import com.kiseki.interaction.validation.ValidationRule;

import lombok.RequiredArgsConstructor;

/**
 * Validates that a user exists in the system.
 */
@Component
@RequiredArgsConstructor
public class UserExistsRule implements ValidationRule<CommentValidationContext> {

    private final IUserValidationClient userValidationClient;

    @Override
    public ValidationResult validate(CommentValidationContext context) {
        if (!userValidationClient.isUserExists(context.getUserId())) {
            return ValidationResult.failure("Invalid user ID");
        }
        return ValidationResult.success();
    }
}
