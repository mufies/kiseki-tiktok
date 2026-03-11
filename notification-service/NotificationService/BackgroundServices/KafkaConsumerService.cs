using Confluent.Kafka;
using Microsoft.AspNetCore.SignalR;
using System.Text.Json;
using NotificationService.Configuration;
using NotificationService.DTOs;
using NotificationService.Hubs;
using NotificationService.Models;
using NotificationService.Services;
using Microsoft.Extensions.Options;

namespace NotificationService.BackgroundServices;

public class KafkaConsumerService : BackgroundService
{
    private readonly IServiceProvider _serviceProvider;
    private readonly KafkaSettings _kafkaSettings;
    private readonly IHubContext<NotificationHub> _hubContext;
    private readonly ILogger<KafkaConsumerService> _logger;

    public KafkaConsumerService(
        IServiceProvider serviceProvider,
        IOptions<KafkaSettings> kafkaSettings,
        IHubContext<NotificationHub> hubContext,
        ILogger<KafkaConsumerService> logger)
    {
        _serviceProvider = serviceProvider;
        _kafkaSettings = kafkaSettings.Value;
        _hubContext = hubContext;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        await Task.Delay(5000, stoppingToken); // Wait for Kafka to be ready

        var config = new ConsumerConfig
        {
            BootstrapServers = _kafkaSettings.BootstrapServers,
            GroupId = _kafkaSettings.GroupId,
            AutoOffsetReset = AutoOffsetReset.Earliest,
            EnableAutoCommit = false
        };

        using var consumer = new ConsumerBuilder<Ignore, string>(config).Build();
        consumer.Subscribe(_kafkaSettings.Topics);

        _logger.LogInformation("Kafka consumer started. Subscribed to topics: {Topics}",
            string.Join(", ", _kafkaSettings.Topics));

        try
        {
            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    var consumeResult = consumer.Consume(stoppingToken);

                    if (consumeResult?.Message?.Value != null)
                    {
                        _logger.LogInformation("Received message from topic {Topic}: {Message}",
                            consumeResult.Topic, consumeResult.Message.Value);

                        await ProcessMessageAsync(consumeResult.Topic, consumeResult.Message.Value);

                        consumer.Commit(consumeResult);
                        _logger.LogInformation("Message committed successfully");
                    }
                }
                catch (ConsumeException ex)
                {
                    _logger.LogError(ex, "Error consuming message from Kafka");
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "Error processing Kafka message");
                }
            }
        }
        finally
        {
            consumer.Close();
            _logger.LogInformation("Kafka consumer stopped");
        }
    }

    private async Task ProcessMessageAsync(string topic, string message)
    {
        try
        {
            var eventDto = JsonSerializer.Deserialize<NotificationEventDto>(message);
            if (eventDto == null)
            {
                _logger.LogWarning("Failed to deserialize message: {Message}", message);
                return;
            }

            NotificationType notificationType = topic switch
            {
                "interaction.liked" => NotificationType.Like,
                "interaction.commented" => NotificationType.Comment,
                "interaction.bookmarked" => NotificationType.Bookmark,
                "user.followed" => NotificationType.Follow,
                _ => throw new InvalidOperationException($"Unknown topic: {topic}")
            };

            var notification = new Notification
            {
                Id = Guid.NewGuid(),
                UserId = eventDto.ToUserId,
                FromUserId = eventDto.FromUserId,
                Type = notificationType,
                VideoId = eventDto.VideoId,
                CommentId = eventDto.CommentId,
                IsRead = false,
                CreatedAt = DateTime.UtcNow
            };

            using var scope = _serviceProvider.CreateScope();
            var notificationService = scope.ServiceProvider.GetRequiredService<INotificationService>();
            var emailService = scope.ServiceProvider.GetRequiredService<IEmailService>();

            var created = await notificationService.CreateNotificationAsync(notification);

            // Send real-time notification via SignalR
            var notificationDto = NotificationDto.FromNotification(created);
            await _hubContext.Clients.Group($"user:{eventDto.ToUserId}")
                .SendAsync("ReceiveNotification", notificationDto);

            _logger.LogInformation("Notification created and sent via SignalR for user {UserId}", eventDto.ToUserId);

            // Send email for FOLLOW notifications
            if (notificationType == NotificationType.Follow)
            {
                // In production, fetch user email from User Service
                // For now, we'll skip email sending or use a placeholder
                _logger.LogInformation("Follow notification created for user {UserId}", eventDto.ToUserId);
                // await emailService.SendFollowNotificationAsync(userEmail, fromUsername);
            }
        }
        catch (JsonException ex)
        {
            _logger.LogError(ex, "Failed to deserialize Kafka message: {Message}", message);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error processing notification event");
        }
    }
}
