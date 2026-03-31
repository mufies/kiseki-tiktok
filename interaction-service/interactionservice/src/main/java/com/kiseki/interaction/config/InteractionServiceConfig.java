package com.kiseki.interaction.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.context.annotation.Primary;

import com.kiseki.interaction.service.InteractionEventPublisher;
import com.kiseki.interaction.service.adapters.InteractionRepositoryAdapter;
import com.kiseki.interaction.service.adapters.UserValidationClientAdapter;
import com.kiseki.interaction.service.adapters.VideoValidationClientAdapter;
import com.kiseki.interaction.service.interfaces.IInteractionEventPublisher;
import com.kiseki.interaction.service.interfaces.IInteractionRepository;
import com.kiseki.interaction.service.interfaces.IUserValidationClient;
import com.kiseki.interaction.service.interfaces.IVideoValidationClient;

/**
 * Configuration for interaction service dependencies.
 */
@Configuration
public class InteractionServiceConfig {

    @Bean
    @Primary
    public IVideoValidationClient videoValidationClient(VideoValidationClientAdapter adapter) {
        return adapter;
    }

    @Bean
    @Primary
    public IUserValidationClient userValidationClient(UserValidationClientAdapter adapter) {
        return adapter;
    }

    @Bean
    @Primary
    public IInteractionEventPublisher eventPublisher(InteractionEventPublisher publisher) {
        return publisher;
    }

    @Bean
    @Primary
    public IInteractionRepository repositoryInterface(InteractionRepositoryAdapter adapter) {
        return adapter;
    }
}
