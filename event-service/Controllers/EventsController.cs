using Microsoft.AspNetCore.Mvc;
using EventService.DTOs;
using EventService.Models;
using EventService.Services;

namespace EventService.Controllers;

[ApiController]
[Route("events")]
public class EventsController(IWatchEventService eventService) : ControllerBase
{
    [HttpPost("watch")]
    public async Task<IActionResult> Watch(
        [FromBody] WatchEventRequest request,
        CancellationToken ct)
    {
        if (!ModelState.IsValid)
            return BadRequest(ModelState);

        var watchEvent = new WatchEvent
        {
            UserId    = request.UserId,
            VideoId   = request.VideoId,
            WatchPct  = request.WatchPct,
            Liked     = request.Liked,
            Timestamp = request.Timestamp?.ToUniversalTime() ?? DateTime.UtcNow
        };

        await eventService.ProcessAsync(watchEvent, ct);

        return Ok(new { message = "Event processed", user_id = request.UserId });
    }
}
