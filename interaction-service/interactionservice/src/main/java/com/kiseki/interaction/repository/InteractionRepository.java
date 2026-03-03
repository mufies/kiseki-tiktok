package com.kiseki.interaction.repository;

import com.kiseki.interaction.entity.Interaction;
import com.kiseki.interaction.entity.InteractionType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Repository
public interface InteractionRepository extends JpaRepository<Interaction, Long> {
    Optional<Interaction> findByUserIdAndVideoIdAndType(UUID userId, UUID videoId, InteractionType type);
    long countByVideoIdAndType(UUID videoId, InteractionType type);
    List<Interaction> findByVideoIdAndTypeOrderByCreatedAtDesc(UUID videoId, InteractionType type);
}
