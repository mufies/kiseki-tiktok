namespace NotificationService.DTOs;

public class NotificationEventDto
{
    public string Type { get; set; } = string.Empty;
    public string FromUserId { get; set; } = string.Empty;
    public string ToUserId { get; set; } = string.Empty;
    public string? VideoId { get; set; }
    public string? CommentId { get; set; }
}
