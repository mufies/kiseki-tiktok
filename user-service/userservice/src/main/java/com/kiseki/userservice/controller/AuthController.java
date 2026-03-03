package com.kiseki.userservice.controller;

import com.kiseki.userservice.dto.request.LoginRequest;
import com.kiseki.userservice.dto.request.RegisterRequest;
import com.kiseki.userservice.dto.response.AuthResponse;
import com.kiseki.userservice.service.AuthService;
import com.kiseki.userservice.utils.RefreshTokenHolder;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/auth")
@RequiredArgsConstructor
public class AuthController {

    private final AuthService authService;
    private static final int COOKIE_MAX_AGE = 30 * 24 * 60 * 60;

    @PostMapping("/register")
    public ResponseEntity<AuthResponse> register(
            @RequestBody RegisterRequest request,
            HttpServletResponse response) {
        AuthResponse auth = authService.register(request);
        setRefreshCookie(response);
        return ResponseEntity.status(HttpStatus.CREATED).body(auth);
    }

    @PostMapping("/login")
    public ResponseEntity<AuthResponse> login(
            @RequestBody LoginRequest request,
            HttpServletResponse response) {
        AuthResponse auth = authService.login(request);
        setRefreshCookie(response);
        return ResponseEntity.ok(auth);
    }

    @PostMapping("/refresh")
    public ResponseEntity<AuthResponse> refresh(
            @CookieValue(name = "refresh_token", required = false) String refreshToken,
            HttpServletResponse response) {
        if (refreshToken == null) {
            return ResponseEntity.status(HttpStatus.UNAUTHORIZED).build();
        }
        AuthResponse auth = authService.refresh(refreshToken);
        setRefreshCookie(response);
        return ResponseEntity.ok(auth);
    }

    @PostMapping("/logout")
    public ResponseEntity<Void> logout(
            @RequestAttribute("userId") String userId,
            HttpServletResponse response) {
        authService.logout(userId);
        clearRefreshCookie(response);
        return ResponseEntity.noContent().build();
    }

    // ------------------------------------------------------------------ helpers

    private void setRefreshCookie(HttpServletResponse response) {
        String token = RefreshTokenHolder.get();
        Cookie cookie = new Cookie("refresh_token", token);
        cookie.setHttpOnly(true);
        cookie.setSecure(false);   // set true in production (HTTPS)
        cookie.setPath("/auth");
        cookie.setMaxAge(COOKIE_MAX_AGE);
        response.addCookie(cookie);
        RefreshTokenHolder.clear();
    }

    private void clearRefreshCookie(HttpServletResponse response) {
        Cookie cookie = new Cookie("refresh_token", "");
        cookie.setHttpOnly(true);
        cookie.setPath("/auth");
        cookie.setMaxAge(0);
        response.addCookie(cookie);
    }
}
