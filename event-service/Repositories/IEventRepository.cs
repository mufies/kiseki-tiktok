using EventService.Models;

namespace EventService.Repositories;

public interface IEventRepository
{
    Task SaveEventAsync(WatchEvent watchEvent, CancellationToken ct = default);
    Task<IReadOnlyList<WatchEvent>> GetUserEventsAsync(string userId, CancellationToken ct = default);
}
