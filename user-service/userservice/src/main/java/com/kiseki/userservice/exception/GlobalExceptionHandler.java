package com.kiseki.userservice.exception;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import java.time.LocalDateTime;
import java.util.Map;

@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(RuntimeException.class)
    public ResponseEntity<Map<String, Object>> handleRuntimeException(RuntimeException ex) {
        String message = ex.getMessage() != null ? ex.getMessage() : "An unexpected error occurred";

        HttpStatus status = switch (message) {
            case "Email already exists"             -> HttpStatus.CONFLICT;           // 409
            case "User not found"                   -> HttpStatus.NOT_FOUND;          // 404
            case "Invalid credentials"              -> HttpStatus.UNAUTHORIZED;       // 401
            case "Invalid refresh token"            -> HttpStatus.UNAUTHORIZED;       // 401
            case "Refresh token expired, please login again" -> HttpStatus.UNAUTHORIZED; // 401
            default                                 -> HttpStatus.INTERNAL_SERVER_ERROR; // 500
        };

        Map<String, Object> body = Map.of(
                "timestamp", LocalDateTime.now().toString(),
                "status",    status.value(),
                "error",     status.getReasonPhrase(),
                "message",   message
        );

        return ResponseEntity.status(status).body(body);
    }
}
