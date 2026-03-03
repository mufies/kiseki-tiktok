package com.kiseki.userservice.repository;

import com.kiseki.userservice.entity.Follow;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;
import java.util.Optional;

public interface FollowRepository extends JpaRepository<Follow, Long> {
    Optional<Follow> findByFollowerIdAndFollowingId(String followerId, String followingId);
    boolean existsByFollowerIdAndFollowingId(String followerId, String followingId);
    int countByFollowerId(String followerId);
    int countByFollowingId(String followingId);
    List<Follow> findByFollowerId(String followerId);
    List<Follow> findByFollowingId(String followingId);
    void deleteByFollowerIdAndFollowingId(String followerId, String followingId);
}
