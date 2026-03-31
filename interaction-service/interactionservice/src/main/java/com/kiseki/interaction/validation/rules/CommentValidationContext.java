package com.kiseki.interaction.validation.rules;

import java.util.UUID;

import lombok.Builder;
import lombok.Getter;

/**
 * Context object for comment validation.
 * Groups all data needed for comment validation rules.
 */
@Getter
@Builder
public class CommentValidationContext {
    private final UUID userId;
    private final UUID videoId;
    private final String content;
}
