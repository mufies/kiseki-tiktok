using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace NotificationService.Models;

[Table("notifications")]
public class Notification
{
    [Key]
    public Guid Id { get; set; }

    [Required]
    public string UserId { get; set; } = string.Empty;

    [Required]
    public string FromUserId { get; set; } = string.Empty;

    [Required]
    public NotificationType Type { get; set; }

    public string? VideoId { get; set; }

    public string? CommentId { get; set; }

    public bool IsRead { get; set; } = false;

    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
}
