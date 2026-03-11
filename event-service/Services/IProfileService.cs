using EventService.Models;

namespace EventService.Services;

public interface IProfileService
{
    Task<UserProfile> RebuildAndCacheAsync(string userId, CancellationToken ct = default);
    Task<UserProfile?> GetFromCacheAsync(string userId, CancellationToken ct = default);
}
