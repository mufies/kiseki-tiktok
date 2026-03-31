package com.kiseki.interaction.service.adapters;

import java.util.List;
import java.util.Optional;
import java.util.UUID;

import org.springframework.stereotype.Component;

import com.kiseki.interaction.entity.Interaction;
import com.kiseki.interaction.entity.InteractionType;
import com.kiseki.interaction.repository.InteractionRepository;
import com.kiseki.interaction.service.interfaces.IInteractionRepository;

import lombok.RequiredArgsConstructor;

/**
 * Adapter that bridges Spring Data JPA InteractionRepository to IInteractionRepository interface.
 */
@Component
@RequiredArgsConstructor
public class InteractionRepositoryAdapter implements IInteractionRepository {

    private final InteractionRepository interactionRepository;

    @Override
    public Optional<Interaction> findByUserIdAndVideoIdAndType(UUID userId, UUID videoId, InteractionType type) {
        return interactionRepository.findByUserIdAndVideoIdAndType(userId, videoId, type);
    }

    @Override
    public void delete(Interaction interaction) {
        interactionRepository.delete(interaction);
    }

    @Override
    public Interaction save(Interaction interaction) {
        return interactionRepository.save(interaction);
    }

    @Override
    public long countByVideoIdAndType(UUID videoId, InteractionType type) {
        return interactionRepository.countByVideoIdAndType(videoId, type);
    }

    @Override
    public List<Interaction> findByVideoIdAndTypeOrderByCreatedAtDesc(UUID videoId, InteractionType type) {
        return interactionRepository.findByVideoIdAndTypeOrderByCreatedAtDesc(videoId, type);
    }

    @Override
    public List<Interaction> findByUserIdAndVideoIdInAndType(UUID userId, List<UUID> videoIds, InteractionType type) {
        return interactionRepository.findByUserIdAndVideoIdInAndType(userId, videoIds, type);
    }

    @Override
    public List<Object[]> countInteractionsByVideoIds(List<UUID> videoIds) {
        return interactionRepository.countInteractionsByVideoIds(videoIds);
    }

    @Override
    public List<Interaction> findByUserIdAndTypeOrderByCreatedAtDesc(UUID userId, InteractionType type) {
        return interactionRepository.findByUserIdAndTypeOrderByCreatedAtDesc(userId, type);
    }
}
