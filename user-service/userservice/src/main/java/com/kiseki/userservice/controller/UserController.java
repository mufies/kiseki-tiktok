package com.kiseki.userservice.controller;

import com.kiseki.userservice.dto.response.UserResponse;
import com.kiseki.userservice.service.UserService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import com.kiseki.userservice.dto.request.UpdateProfileRequest;
import java.util.List;

@RestController
@RequestMapping("/api/users")
@RequiredArgsConstructor
public class UserController {

    private final UserService userService;

    @GetMapping("/me")
    public ResponseEntity<UserResponse> getCurrentUser(@RequestAttribute("userId") String userId) {
        return ResponseEntity.ok(userService.getUserProfile(userId));
    }

    @PutMapping("/me")
    public ResponseEntity<UserResponse> updateProfile(
            @RequestAttribute("userId") String userId,
            @RequestBody UpdateProfileRequest request) {
        return ResponseEntity.ok(userService.updateProfile(userId, request));
    }

    @GetMapping("/{id}")
    public ResponseEntity<UserResponse> getUserById(@PathVariable String id) {
        return ResponseEntity.ok(userService.getUserById(id));
    }

    @PostMapping("/{id}/follow")
    public ResponseEntity<?> followUser(
            @RequestAttribute("userId") String followerId,
            @PathVariable("id") String followingId) {
        userService.followUser(followerId, followingId);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/{id}/follow")
    public ResponseEntity<?> unfollowUser(
            @RequestAttribute("userId") String followerId,
            @PathVariable("id") String followingId) {
        userService.unfollowUser(followerId, followingId);
        return ResponseEntity.ok().build();
    }

    @GetMapping("/{id}/followers")
    public ResponseEntity<List<UserResponse>> getFollowers(@PathVariable("id") String id) {
        return ResponseEntity.ok(userService.getFollowers(id));
    }

    @GetMapping("/{id}/following")
    public ResponseEntity<List<UserResponse>> getFollowing(@PathVariable("id") String id) {
        return ResponseEntity.ok(userService.getFollowing(id));
    }
}
