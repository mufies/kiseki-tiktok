using Microsoft.AspNetCore.Mvc;
using EventService.Models;
using EventService.Services;
using EventService.Repositories;

namespace EventService.Controllers;

[ApiController]
[Route("profile")]
public class ProfileController(
    IProfileService profileService,
    IEventRepository eventRepo) : ControllerBase
{
    [HttpGet("{userId}")]
    public async Task<IActionResult> GetProfile(string userId, CancellationToken ct)
    {
        var profile = await profileService.GetFromCacheAsync(userId, ct);
        if (profile is not null)
            return Ok(profile);

        var events = await eventRepo.GetUserEventsAsync(userId, ct);
        if (events.Count > 0)
        {
            profile = await profileService.RebuildAndCacheAsync(userId, ct);
            return Ok(profile);
        }

        return Ok(new UserProfile
        {
            UserId     = userId,
            Categories = [],
            Hashtags   = [],
            UpdatedAt  = DateTime.UtcNow
        });
    }
}
