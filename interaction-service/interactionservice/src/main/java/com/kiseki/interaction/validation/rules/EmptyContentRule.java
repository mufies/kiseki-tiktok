package com.kiseki.interaction.validation.rules;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.validation.ValidationResult;
import com.kiseki.interaction.validation.ValidationRule;

/**
 * Validates that comment content is not empty or null.
 */
@Component
public class EmptyContentRule implements ValidationRule<CommentValidationContext> {

    @Override
    public ValidationResult validate(CommentValidationContext context) {
        if (context.getContent() == null || context.getContent().trim().isEmpty()) {
            return ValidationResult.failure("Comment content cannot be empty");
        }
        return ValidationResult.success();
    }
}
