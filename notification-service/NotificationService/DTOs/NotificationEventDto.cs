using System.Text.Json.Serialization;

namespace NotificationService.DTOs;

public class NotificationEventDto
{
    [JsonPropertyName("type")]
    public string Type { get; set; } = string.Empty;
    
    [JsonPropertyName("fromUserId")]
    public string FromUserId { get; set; } = string.Empty;
    
    [JsonPropertyName("toUserId")]
    public string ToUserId { get; set; } = string.Empty;
    
    [JsonPropertyName("videoId")]
    public string? VideoId { get; set; }
    
    [JsonPropertyName("commentId")]
    public string? CommentId { get; set; }
}
