package com.kiseki.interaction.validation.rules;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.util.UUID;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.kiseki.interaction.service.interfaces.IUserValidationClient;
import com.kiseki.interaction.service.interfaces.IVideoValidationClient;
import com.kiseki.interaction.validation.ValidationResult;

/**
 * Unit tests for comment validation rules.
 * Demonstrates Single Responsibility Principle - each rule tested independently.
 * Demonstrates Open/Closed Principle - new rules can be added without modifying existing tests.
 */
@ExtendWith(MockitoExtension.class)
@DisplayName("Comment Validation Rules Unit Tests")
class CommentValidationRulesTest {

    @Mock
    private IUserValidationClient userValidationClient;

    @Mock
    private IVideoValidationClient videoValidationClient;

    private EmptyContentRule emptyContentRule;
    private CommentLengthRule commentLengthRule;
    private UserExistsRule userExistsRule;
    private VideoExistsRule videoExistsRule;

    private UUID testUserId;
    private UUID testVideoId;

    @BeforeEach
    void setUp() {
        emptyContentRule = new EmptyContentRule();
        commentLengthRule = new CommentLengthRule();
        userExistsRule = new UserExistsRule(userValidationClient);
        videoExistsRule = new VideoExistsRule(videoValidationClient);

        testUserId = UUID.randomUUID();
        testVideoId = UUID.randomUUID();
    }

    // EmptyContentRule Tests

    @Test
    @DisplayName("EmptyContentRule: Should fail when content is null")
    void emptyContentRule_ShouldFailWhenNull() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .content(null)
            .build();

        // Act
        ValidationResult result = emptyContentRule.validate(context);

        // Assert
        assertFalse(result.isValid());
        assertEquals("Comment content cannot be empty", result.getErrorMessage());
    }

    @Test
    @DisplayName("EmptyContentRule: Should fail when content is empty string")
    void emptyContentRule_ShouldFailWhenEmpty() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .content("")
            .build();

        // Act
        ValidationResult result = emptyContentRule.validate(context);

        // Assert
        assertFalse(result.isValid());
        assertEquals("Comment content cannot be empty", result.getErrorMessage());
    }

    @Test
    @DisplayName("EmptyContentRule: Should fail when content is only whitespace")
    void emptyContentRule_ShouldFailWhenOnlyWhitespace() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .content("   ")
            .build();

        // Act
        ValidationResult result = emptyContentRule.validate(context);

        // Assert
        assertFalse(result.isValid());
    }

    @Test
    @DisplayName("EmptyContentRule: Should pass when content is valid")
    void emptyContentRule_ShouldPassWhenValid() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .content("Great video!")
            .build();

        // Act
        ValidationResult result = emptyContentRule.validate(context);

        // Assert
        assertTrue(result.isValid());
    }

    // CommentLengthRule Tests

    @Test
    @DisplayName("CommentLengthRule: Should fail when content exceeds max length")
    void commentLengthRule_ShouldFailWhenTooLong() {
        // Arrange
        String longContent = "a".repeat(1001);
        CommentValidationContext context = CommentValidationContext.builder()
            .content(longContent)
            .build();

        // Act
        ValidationResult result = commentLengthRule.validate(context);

        // Assert
        assertFalse(result.isValid());
        assertTrue(result.getErrorMessage().contains("cannot exceed 1000 characters"));
    }

    @Test
    @DisplayName("CommentLengthRule: Should fail when trimmed content is too short")
    void commentLengthRule_ShouldFailWhenTooShort() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .content("   ")
            .build();

        // Act
        ValidationResult result = commentLengthRule.validate(context);

        // Assert
        assertFalse(result.isValid());
        assertEquals("Comment content too short", result.getErrorMessage());
    }

    @Test
    @DisplayName("CommentLengthRule: Should pass when content length is valid")
    void commentLengthRule_ShouldPassWhenValid() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .content("This is a valid comment!")
            .build();

        // Act
        ValidationResult result = commentLengthRule.validate(context);

        // Assert
        assertTrue(result.isValid());
    }

    @Test
    @DisplayName("CommentLengthRule: Should pass when content is exactly max length")
    void commentLengthRule_ShouldPassAtMaxLength() {
        // Arrange
        String maxContent = "a".repeat(1000);
        CommentValidationContext context = CommentValidationContext.builder()
            .content(maxContent)
            .build();

        // Act
        ValidationResult result = commentLengthRule.validate(context);

        // Assert
        assertTrue(result.isValid());
    }

    // UserExistsRule Tests

    @Test
    @DisplayName("UserExistsRule: Should fail when user doesn't exist")
    void userExistsRule_ShouldFailWhenUserDoesNotExist() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .userId(testUserId)
            .build();
        when(userValidationClient.isUserExists(testUserId)).thenReturn(false);

        // Act
        ValidationResult result = userExistsRule.validate(context);

        // Assert
        assertFalse(result.isValid());
        assertEquals("Invalid user ID", result.getErrorMessage());
    }

    @Test
    @DisplayName("UserExistsRule: Should pass when user exists")
    void userExistsRule_ShouldPassWhenUserExists() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .userId(testUserId)
            .build();
        when(userValidationClient.isUserExists(testUserId)).thenReturn(true);

        // Act
        ValidationResult result = userExistsRule.validate(context);

        // Assert
        assertTrue(result.isValid());
    }

    // VideoExistsRule Tests

    @Test
    @DisplayName("VideoExistsRule: Should fail when video doesn't exist")
    void videoExistsRule_ShouldFailWhenVideoDoesNotExist() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .videoId(testVideoId)
            .build();
        when(videoValidationClient.validateVideoExists(testVideoId)).thenReturn(false);

        // Act
        ValidationResult result = videoExistsRule.validate(context);

        // Assert
        assertFalse(result.isValid());
        assertEquals("Invalid video ID", result.getErrorMessage());
    }

    @Test
    @DisplayName("VideoExistsRule: Should pass when video exists")
    void videoExistsRule_ShouldPassWhenVideoExists() {
        // Arrange
        CommentValidationContext context = CommentValidationContext.builder()
            .videoId(testVideoId)
            .build();
        when(videoValidationClient.validateVideoExists(testVideoId)).thenReturn(true);

        // Act
        ValidationResult result = videoExistsRule.validate(context);

        // Assert
        assertTrue(result.isValid());
    }
}
