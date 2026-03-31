using NotificationService.DTOs;
using NotificationService.Models;

namespace NotificationService.Services;

public interface INotificationService
{
    Task<Notification> CreateNotificationAsync(Notification notification);
    Task<PagedResult<NotificationDto>> GetNotificationsAsync(string userId, int page, int pageSize);
    Task<PagedResult<NotificationDetailDto>> GetNotificationDetailsAsync(string userId, int page, int pageSize);
    Task MarkAsReadAsync(string userId, List<Guid> notificationIds);
    Task MarkAllAsReadAsync(string userId);
    Task<int> GetUnreadCountAsync(string userId);
}
