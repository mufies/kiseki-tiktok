using System.ComponentModel.DataAnnotations;

namespace EventService.DTOs;

public record WatchEventRequest(
    [Required] string UserId,
    [Required] string VideoId,
    [Range(0, 100)] float WatchPct,
    bool Liked,
    DateTime? Timestamp
);
