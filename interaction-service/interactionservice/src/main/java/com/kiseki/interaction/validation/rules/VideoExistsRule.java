package com.kiseki.interaction.validation.rules;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.service.interfaces.IVideoValidationClient;
import com.kiseki.interaction.validation.ValidationResult;
import com.kiseki.interaction.validation.ValidationRule;

import lombok.RequiredArgsConstructor;

/**
 * Validates that a video exists in the system.
 */
@Component
@RequiredArgsConstructor
public class VideoExistsRule implements ValidationRule<CommentValidationContext> {

    private final IVideoValidationClient videoValidationClient;

    @Override
    public ValidationResult validate(CommentValidationContext context) {
        if (!videoValidationClient.validateVideoExists(context.getVideoId())) {
            return ValidationResult.failure("Invalid video ID");
        }
        return ValidationResult.success();
    }
}
