using NotificationService.Models;

namespace NotificationService.DTOs;

public class NotificationDto
{
    public Guid Id { get; set; }
    public string UserId { get; set; } = string.Empty;
    public string FromUserId { get; set; } = string.Empty;
    public NotificationType Type { get; set; }
    public string? VideoId { get; set; }
    public string? CommentId { get; set; }
    public bool IsRead { get; set; }
    public DateTime CreatedAt { get; set; }

    public static NotificationDto FromNotification(Notification notification)
    {
        return new NotificationDto
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
    }
}
