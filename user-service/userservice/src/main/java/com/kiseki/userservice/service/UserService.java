package com.kiseki.userservice.service;

import com.kiseki.userservice.dto.response.UserResponse;
import com.kiseki.userservice.dto.request.UpdateProfileRequest;
import com.kiseki.userservice.entity.User;
import com.kiseki.userservice.entity.Follow;
import com.kiseki.userservice.kafka.KafkaProducerService;
import com.kiseki.userservice.repository.UserRepository;
import com.kiseki.userservice.repository.FollowRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class UserService {

    private final UserRepository userRepository;
    private final FollowRepository followRepository;
    private final KafkaProducerService kafkaProducerService;

    public UserResponse getUserProfile(String userId) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("User not found"));
        return mapToResponse(user);
    }

    public UserResponse getUserById(String id) {
        User user = userRepository.findById(id)
                .orElseThrow(() -> new RuntimeException("User not found"));
        return mapToResponse(user);
    }

    @Transactional
    public UserResponse updateProfile(String userId, UpdateProfileRequest request) {
        User user = userRepository.findById(userId)
                .orElseThrow(() -> new RuntimeException("User not found"));
        if (request.getUsername() != null) user.setUsername(request.getUsername());
        if (request.getAvatarUrl() != null) user.setAvatarUrl(request.getAvatarUrl());
        return mapToResponse(userRepository.save(user));
    }

    @Transactional
    public void followUser(String followerId, String followingId) {
        if (followerId.equals(followingId)) {
            throw new RuntimeException("Cannot follow yourself");
        }
        if (!userRepository.existsById(followingId)) {
            throw new RuntimeException("Target user not found");
        }
        if (!followRepository.existsByFollowerIdAndFollowingId(followerId, followingId)) {
            followRepository.save(Follow.builder()
                    .followerId(followerId)
                    .followingId(followingId)
                    .build());

            // Send follow notification event
            kafkaProducerService.sendFollowEvent(followerId, followingId);
        }
    }

    @Transactional
    public void unfollowUser(String followerId, String followingId) {
        followRepository.deleteByFollowerIdAndFollowingId(followerId, followingId);
    }

    public List<UserResponse> getFollowers(String userId) {
        return followRepository.findByFollowingId(userId).stream()
                .map(follow -> getUserById(follow.getFollowerId()))
                .collect(Collectors.toList());
    }

    public List<UserResponse> getFollowing(String userId) {
        return followRepository.findByFollowerId(userId).stream()
                .map(follow -> getUserById(follow.getFollowingId()))
                .collect(Collectors.toList());
    }

    private UserResponse mapToResponse(User user) {
        int followerCount = followRepository.countByFollowingId(user.getId());
        int followingCount = followRepository.countByFollowerId(user.getId());
        return UserResponse.builder()
                .id(user.getId())
                .email(user.getEmail())
                .username(user.getUsername())
                .avatarUrl(user.getAvatarUrl())
                .followerCount(followerCount)
                .followingCount(followingCount)
                .createdAt(user.getCreatedAt())
                .build();
    }
}
