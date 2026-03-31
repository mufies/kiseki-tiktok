using StackExchange.Redis;
using NotificationService.DTOs;
using NotificationService.Models;
using NotificationService.Repositories;

namespace NotificationService.Services;

public class NotificationService : INotificationService
{
    private readonly INotificationRepository _repository;
    private readonly IConnectionMultiplexer _redis;
    private readonly IExternalServiceClient _externalServiceClient;
    private readonly ILogger<NotificationService> _logger;

    public NotificationService(
        INotificationRepository repository,
        IConnectionMultiplexer redis,
        IExternalServiceClient externalServiceClient,
        ILogger<NotificationService> logger)
    {
        _repository = repository;
        _redis = redis;
        _externalServiceClient = externalServiceClient;
        _logger = logger;
    }

    public async Task<Notification> CreateNotificationAsync(Notification notification)
    {
        // Validate notification data
        if (string.IsNullOrWhiteSpace(notification.UserId) ||
            string.IsNullOrWhiteSpace(notification.FromUserId))
        {
            throw new ArgumentException("UserId and FromUserId are required");
        }

        // Prevent self-notifications (fraud prevention)
        if (notification.UserId.Equals(notification.FromUserId, StringComparison.OrdinalIgnoreCase))
        {
            _logger.LogWarning("Attempted self-notification blocked: UserId={UserId}", notification.UserId);
            throw new InvalidOperationException("Cannot create notification for self");
        }

        var created = await _repository.CreateAsync(notification);

        // Update Redis cache - increment unread count with expiration
        var db = _redis.GetDatabase();
        var cacheKey = $"notification:unread:{SanitizeUserId(notification.UserId)}";
        await db.StringIncrementAsync(cacheKey);
        await db.KeyExpireAsync(cacheKey, TimeSpan.FromHours(24)); // Add cache expiration

        return created;
    }

    private static string SanitizeUserId(string userId)
    {
        // Remove any characters that could cause Redis key injection
        if (string.IsNullOrWhiteSpace(userId))
            throw new ArgumentException("UserId cannot be null or empty");

        // Only allow alphanumeric, hyphens, and underscores
        return System.Text.RegularExpressions.Regex.Replace(userId, @"[^a-zA-Z0-9\-_]", "");
    }

    public async Task<PagedResult<NotificationDto>> GetNotificationsAsync(string userId, int page, int pageSize)
    {
        // Input validation (defense in depth)
        if (string.IsNullOrWhiteSpace(userId))
            throw new ArgumentException("UserId is required");

        if (page < 1 || page > 1000)
            throw new ArgumentException("Page must be between 1 and 1000");

        if (pageSize < 1 || pageSize > 100)
            throw new ArgumentException("PageSize must be between 1 and 100");

        var skip = (page - 1) * pageSize;
        var notifications = await _repository.GetByUserIdAsync(userId, skip, pageSize);
        var totalCount = await _repository.GetTotalCountAsync(userId);
        var totalPages = (int)Math.Ceiling(totalCount / (double)pageSize);

        return new PagedResult<NotificationDto>
        {
            Items = notifications.Select(NotificationDto.FromNotification).ToList(),
            TotalCount = totalCount,
            Page = page,
            PageSize = pageSize,
            TotalPages = totalPages
        };
    }

    public async Task MarkAsReadAsync(string userId, List<Guid> notificationIds)
    {
        if (string.IsNullOrWhiteSpace(userId))
            throw new ArgumentException("UserId is required");

        if (notificationIds == null || notificationIds.Count == 0)
            throw new ArgumentException("NotificationIds cannot be empty");

        await _repository.MarkAsReadAsync(userId, notificationIds);

        // Invalidate cache - force recalculation on next request
        var db = _redis.GetDatabase();
        var cacheKey = $"notification:unread:{SanitizeUserId(userId)}";
        await db.KeyDeleteAsync(cacheKey);
    }

    public async Task MarkAllAsReadAsync(string userId)
    {
        if (string.IsNullOrWhiteSpace(userId))
            throw new ArgumentException("UserId is required");

        await _repository.MarkAllAsReadAsync(userId);

        // Update cache to 0 with expiration
        var db = _redis.GetDatabase();
        var cacheKey = $"notification:unread:{SanitizeUserId(userId)}";
        await db.StringSetAsync(cacheKey, 0, TimeSpan.FromHours(24));
    }

    public async Task<int> GetUnreadCountAsync(string userId)
    {
        if (string.IsNullOrWhiteSpace(userId))
            throw new ArgumentException("UserId is required");

        var db = _redis.GetDatabase();
        var cacheKey = $"notification:unread:{SanitizeUserId(userId)}";

        var cachedCount = await db.StringGetAsync(cacheKey);
        if (cachedCount.HasValue)
        {
            return (int)cachedCount;
        }

        // Cache miss - fetch from database
        var count = await _repository.GetUnreadCountAsync(userId);
        await db.StringSetAsync(cacheKey, count, TimeSpan.FromHours(24)); // Add cache expiration

        return count;
    }

    public async Task<PagedResult<NotificationDetailDto>> GetNotificationDetailsAsync(string userId, int page, int pageSize)
    {
        // Input validation (defense in depth)
        if (string.IsNullOrWhiteSpace(userId))
            throw new ArgumentException("UserId is required");

        if (page < 1 || page > 1000)
            throw new ArgumentException("Page must be between 1 and 1000");

        if (pageSize < 1 || pageSize > 100)
            throw new ArgumentException("PageSize must be between 1 and 100");

        var skip = (page - 1) * pageSize;
        var notifications = await _repository.GetByUserIdAsync(userId, skip, pageSize);
        var totalCount = await _repository.GetTotalCountAsync(userId);
        var totalPages = (int)Math.Ceiling(totalCount / (double)pageSize);

        // Enrich notifications with user and video data
        var detailedNotifications = new List<NotificationDetailDto>();

        foreach (var notification in notifications)
        {
            var detail = new NotificationDetailDto
            {
                Id = notification.Id,
                UserId = notification.UserId,
                FromUserId = notification.FromUserId,
                Type = notification.Type,
                VideoId = notification.VideoId,
                CommentId = notification.CommentId,
                IsRead = notification.IsRead,
                CreatedAt = notification.CreatedAt
            };

            // Fetch user profile
            var userProfile = await _externalServiceClient.GetUserProfileAsync(notification.FromUserId);
            if (userProfile != null)
            {
                detail.FromUsername = userProfile.Username;
                detail.FromAvatarUrl = userProfile.AvatarUrl;
            }
            else
            {
                detail.FromUsername = "Unknown User";
            }

            // Fetch video info if applicable
            if (!string.IsNullOrEmpty(notification.VideoId))
            {
                var videoInfo = await _externalServiceClient.GetVideoInfoAsync(notification.VideoId);
                if (videoInfo != null)
                {
                    detail.VideoTitle = videoInfo.Title;
                    detail.VideoThumbnail = videoInfo.VideoThumbnail;
                }
            }

            // Fetch comment info if applicable
            if (!string.IsNullOrEmpty(notification.CommentId))
            {
                var commentInfo = await _externalServiceClient.GetCommentInfoAsync(notification.CommentId);
                if (commentInfo != null)
                {
                    detail.CommentContent = commentInfo.Content;
                }
            }

            // Generate message based on type
            detail.Message = GenerateNotificationMessage(detail);

            detailedNotifications.Add(detail);
        }

        return new PagedResult<NotificationDetailDto>
        {
            Items = detailedNotifications,
            TotalCount = totalCount,
            Page = page,
            PageSize = pageSize,
            TotalPages = totalPages
        };
    }

    private static string GenerateNotificationMessage(NotificationDetailDto notification)
    {
        return notification.Type switch
        {
            NotificationType.Like => "liked your video",
            NotificationType.Comment => !string.IsNullOrEmpty(notification.CommentContent)
                ? $"commented: {notification.CommentContent}"
                : "commented on your video",
            NotificationType.Follow => "started following you",
            NotificationType.Bookmark => "bookmarked your video",
            _ => "interacted with your content"
        };
    }
}
