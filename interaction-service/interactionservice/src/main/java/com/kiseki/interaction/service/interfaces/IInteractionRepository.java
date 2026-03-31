package com.kiseki.interaction.service.interfaces;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import com.kiseki.interaction.entity.Interaction;
import com.kiseki.interaction.entity.InteractionType;

/**
 * Repository interface for interaction operations.
 */
public interface IInteractionRepository {

    Optional<Interaction> findByUserIdAndVideoIdAndType(UUID userId, UUID videoId, InteractionType type);

    void delete(Interaction interaction);

    Interaction save(Interaction interaction);

    long countByVideoIdAndType(UUID videoId, InteractionType type);

    List<Interaction> findByVideoIdAndTypeOrderByCreatedAtDesc(UUID videoId, InteractionType type);

    List<Interaction> findByUserIdAndVideoIdInAndType(UUID userId, List<UUID> videoIds, InteractionType type);

    List<Object[]> countInteractionsByVideoIds(List<UUID> videoIds);

    List<Interaction> findByUserIdAndTypeOrderByCreatedAtDesc(UUID userId, InteractionType type);
}
