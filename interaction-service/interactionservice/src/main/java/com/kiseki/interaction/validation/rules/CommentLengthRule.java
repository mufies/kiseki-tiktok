package com.kiseki.interaction.validation.rules;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.validation.ValidationResult;
import com.kiseki.interaction.validation.ValidationRule;

/**
 * Validates that comment content length is within acceptable limits.
 */
@Component
public class CommentLengthRule implements ValidationRule<CommentValidationContext> {

    private static final int MAX_COMMENT_LENGTH = 1000;
    private static final int MIN_COMMENT_LENGTH = 1;

    @Override
    public ValidationResult validate(CommentValidationContext context) {
        String content = context.getContent();

        if (content == null) {
            return ValidationResult.success();
        }

        if (content.length() > MAX_COMMENT_LENGTH) {
            return ValidationResult.failure(
                String.format("Comment content cannot exceed %d characters", MAX_COMMENT_LENGTH));
        }

        if (content.trim().length() < MIN_COMMENT_LENGTH) {
            return ValidationResult.failure("Comment content too short");
        }

        return ValidationResult.success();
    }
}
