namespace EventService.Models;

public class UserProfile
{
    public string UserId { get; set; } = string.Empty;
    public Dictionary<string, double> Categories { get; set; } = [];
    public Dictionary<string, double> Hashtags { get; set; } = [];
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
