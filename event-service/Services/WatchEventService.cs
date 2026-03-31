using EventService.Models;
using EventService.Repositories;

namespace EventService.Services;

public class WatchEventService(
    IEventRepository eventRepo,
    IProfileService profileService,
    ILogger<WatchEventService> logger) : IWatchEventService
{
    public async Task ProcessAsync(WatchEvent watchEvent, CancellationToken ct = default)
    {
        await eventRepo.SaveEventAsync(watchEvent, ct);
        logger.LogInformation("Saved watch event: user={UserId}, video={VideoId}", watchEvent.UserId, watchEvent.VideoId);

        await profileService.RebuildAndCacheAsync(watchEvent.UserId, ct);
    }
}
