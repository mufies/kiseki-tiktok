using NotificationService.Models;

namespace NotificationService.Repositories;

public interface INotificationRepository
{
    Task<Notification> CreateAsync(Notification notification);
    Task<List<Notification>> GetByUserIdAsync(string userId, int skip, int take);
    Task<int> GetTotalCountAsync(string userId);
    Task<int> GetUnreadCountAsync(string userId);
    Task MarkAsReadAsync(string userId, List<Guid> notificationIds);
    Task MarkAllAsReadAsync(string userId);
    Task<Notification?> GetByIdAsync(Guid id);
}
