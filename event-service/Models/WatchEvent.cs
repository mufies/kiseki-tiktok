using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace EventService.Models;

[Table("watch_events")]
public class WatchEvent
{
    [Key]
    [Column("id")]
    public long Id { get; set; }

    [Required]
    [Column("user_id")]
    public string UserId { get; set; } = string.Empty;

    [Required]
    [Column("video_id")]
    public string VideoId { get; set; } = string.Empty;

    [Column("watch_pct")]
    public float WatchPct { get; set; }

    [Column("liked")]
    public bool Liked { get; set; }

    [Column("timestamp")]
    public DateTime Timestamp { get; set; } = DateTime.UtcNow;
}
