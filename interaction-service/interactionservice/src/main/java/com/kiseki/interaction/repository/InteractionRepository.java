package com.kiseki.interaction.repository;

import com.kiseki.interaction.entity.Interaction;
import com.kiseki.interaction.entity.InteractionType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface InteractionRepository extends JpaRepository<Interaction, Long> {
    Optional<Interaction> findByUserIdAndVideoIdAndType(UUID userId, UUID videoId, InteractionType type);
    long countByVideoIdAndType(UUID videoId, InteractionType type);
    List<Interaction> findByVideoIdAndTypeOrderByCreatedAtDesc(UUID videoId, InteractionType type);

    // Bulk queries for multiple videos
    List<Interaction> findByVideoIdInAndType(List<UUID> videoIds, InteractionType type);
    List<Interaction> findByUserIdAndVideoIdInAndType(UUID userId, List<UUID> videoIds, InteractionType type);

    @Query("SELECT i.videoId, i.type, COUNT(i) FROM Interaction i WHERE i.videoId IN :videoIds GROUP BY i.videoId, i.type")
    List<Object[]> countInteractionsByVideoIds(@Param("videoIds") List<UUID> videoIds);

    // Get user's interactions by type, ordered by most recent
    List<Interaction> findByUserIdAndTypeOrderByCreatedAtDesc(UUID userId, InteractionType type);
}
