using StackExchange.Redis;
using NotificationService.DTOs;
using NotificationService.Models;
using NotificationService.Repositories;

namespace NotificationService.Services;

public class NotificationService : INotificationService
{
    private readonly INotificationRepository _repository;
    private readonly IConnectionMultiplexer _redis;
    private readonly ILogger<NotificationService> _logger;

    public NotificationService(
        INotificationRepository repository,
        IConnectionMultiplexer redis,
        ILogger<NotificationService> logger)
    {
        _repository = repository;
        _redis = redis;
        _logger = logger;
    }

    public async Task<Notification> CreateNotificationAsync(Notification notification)
    {
        var created = await _repository.CreateAsync(notification);

        // Update Redis cache - increment unread count
        var db = _redis.GetDatabase();
        var cacheKey = $"notification:unread:{notification.UserId}";
        await db.StringIncrementAsync(cacheKey);

        return created;
    }

    public async Task<PagedResult<NotificationDto>> GetNotificationsAsync(string userId, int page, int pageSize)
    {
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
        await _repository.MarkAsReadAsync(userId, notificationIds);

        // Invalidate cache - force recalculation on next request
        var db = _redis.GetDatabase();
        var cacheKey = $"notification:unread:{userId}";
        await db.KeyDeleteAsync(cacheKey);
    }

    public async Task MarkAllAsReadAsync(string userId)
    {
        await _repository.MarkAllAsReadAsync(userId);

        // Update cache to 0
        var db = _redis.GetDatabase();
        var cacheKey = $"notification:unread:{userId}";
        await db.StringSetAsync(cacheKey, 0);
    }

    public async Task<int> GetUnreadCountAsync(string userId)
    {
        var db = _redis.GetDatabase();
        var cacheKey = $"notification:unread:{userId}";

        var cachedCount = await db.StringGetAsync(cacheKey);
        if (cachedCount.HasValue)
        {
            return (int)cachedCount;
        }

        // Cache miss - fetch from database
        var count = await _repository.GetUnreadCountAsync(userId);
        await db.StringSetAsync(cacheKey, count);

        return count;
    }
}
