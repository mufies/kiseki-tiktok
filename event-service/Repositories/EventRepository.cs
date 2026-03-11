using Microsoft.EntityFrameworkCore;
using EventService.Data;
using EventService.Models;

namespace EventService.Repositories;

public class EventRepository(AppDbContext db) : IEventRepository
{
    public async Task SaveEventAsync(WatchEvent watchEvent, CancellationToken ct = default)
    {
        db.WatchEvents.Add(watchEvent);
        await db.SaveChangesAsync(ct);
    }

    public async Task<IReadOnlyList<WatchEvent>> GetUserEventsAsync(string userId, CancellationToken ct = default)
    {
        return await db.WatchEvents
            .Where(e => e.UserId == userId)
            .OrderByDescending(e => e.Timestamp)
            .AsNoTracking()
            .ToListAsync(ct);
    }
}
