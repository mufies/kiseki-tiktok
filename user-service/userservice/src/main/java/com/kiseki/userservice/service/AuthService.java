package com.kiseki.userservice.service;

import com.kiseki.userservice.dto.request.LoginRequest;
import com.kiseki.userservice.dto.request.RegisterRequest;
import com.kiseki.userservice.dto.response.AuthResponse;
import com.kiseki.userservice.entity.RefreshToken;
import com.kiseki.userservice.entity.User;
import com.kiseki.userservice.repository.RefreshTokenRepository;
import com.kiseki.userservice.repository.UserRepository;
import com.kiseki.userservice.utils.JwtUtil;
import com.kiseki.userservice.utils.PasswordEncoder;
import com.kiseki.userservice.utils.RefreshTokenHolder;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.UUID;
import org.springframework.beans.factory.annotation.Value;

@Service
@RequiredArgsConstructor
@Transactional
public class AuthService {

    private final UserRepository userRepo;
    private final RefreshTokenRepository refreshRepo;
    private final JwtUtil jwtUtil;
    private final PasswordEncoder passwordEncoder;

    @Value("${jwt.refresh-token-expiry}")
    private long refreshTokenExpiryDays;

    public AuthResponse register(RegisterRequest request) {
        if (userRepo.existsByEmail(request.getEmail())) {
            throw new RuntimeException("Email already exists");
        }
        User user = User.builder()
                .email(request.getEmail())
                .password(passwordEncoder.encode(request.getPassword()))
                .username(request.getUsername())
                .build();
        userRepo.save(user);
        return buildAuthResponse(user);
    }

    public AuthResponse login(LoginRequest request) {
        User user;
        if(request.getEmail() != null){
             user = userRepo.findByEmail(request.getEmail())
                    .orElseThrow(() -> new RuntimeException("User not found"));
        }
        else{
            user = userRepo.findByUsername(request.getUsername())
                .orElseThrow(() -> new RuntimeException("User not found"));
        }
       if (!passwordEncoder.matches(request.getPassword(), user.getPassword())) {
            throw new RuntimeException("Invalid credentials");
        }
        return buildAuthResponse(user);
    }

    public AuthResponse refresh(String rawRefreshToken) {
        RefreshToken stored = refreshRepo
                .findByTokenAndRevokedFalse(rawRefreshToken)
                .orElseThrow(() -> new RuntimeException("Invalid refresh token"));

        if (stored.getExpiresAt().isBefore(LocalDateTime.now())) {
            stored.setRevoked(true);
            refreshRepo.save(stored);
            throw new RuntimeException("Refresh token expired, please login again");
        }

        // Token rotation: revoke old, create new
        stored.setRevoked(true);
        refreshRepo.save(stored);

        return buildAuthResponse(stored.getUser());
    }

    public void logout(String userId) {
        refreshRepo.revokeAllByUserId(userId);
    }

    private AuthResponse buildAuthResponse(User user) {
        String accessToken = jwtUtil.generateAccessToken(user.getId(), user.getEmail());
        String refreshToken = createRefreshToken(user);
        RefreshTokenHolder.set(refreshToken);
        return new AuthResponse(accessToken);
    }

    private String createRefreshToken(User user) {
        String token = UUID.randomUUID().toString();
        RefreshToken rt = RefreshToken.builder()
                .user(user)
                .token(token)
                .expiresAt(LocalDateTime.now().plusDays(refreshTokenExpiryDays))
                .build();
        refreshRepo.save(rt);
        return token;
    }
}
