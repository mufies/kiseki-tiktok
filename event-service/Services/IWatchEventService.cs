using EventService.Models;

namespace EventService.Services;

public interface IWatchEventService
{
    Task ProcessAsync(WatchEvent watchEvent, CancellationToken ct = default);
}
