using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace EventService.Models;

[Table("videos")]
public class Video
{
    [Key]
    [Column("video_id")]
    public string VideoId { get; set; } = string.Empty;

    [Column("title")]
    public string Title { get; set; } = string.Empty;

    [Column("categories")]
    public string[] Categories { get; set; } = [];

    [Column("hashtags")]
    public string[] Hashtags { get; set; } = [];
}
