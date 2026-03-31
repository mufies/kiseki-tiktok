using NotificationService.Models;

namespace NotificationService.DTOs;

public class NotificationDetailDto
{
    public Guid Id { get; set; }
    public string UserId { get; set; } = string.Empty;
    public NotificationType Type { get; set; }
    public bool IsRead { get; set; }
    public DateTime CreatedAt { get; set; }

    // From User Info
    public string FromUserId { get; set; } = string.Empty;
    public string FromUsername { get; set; } = string.Empty;
    public string? FromAvatarUrl { get; set; }

    // Video Info (if applicable)
    public string? VideoId { get; set; }
    public string? VideoTitle { get; set; }
    public string? VideoThumbnail { get; set; }

    // Comment Info (if applicable)
    public string? CommentId { get; set; }
    public string? CommentContent { get; set; }

    // Computed message for display
    public string Message { get; set; } = string.Empty;
}
