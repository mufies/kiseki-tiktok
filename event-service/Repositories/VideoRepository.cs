using Microsoft.EntityFrameworkCore;
using EventService.Data;
using EventService.Models;

namespace EventService.Repositories;

public class VideoRepository(AppDbContext db) : IVideoRepository
{
    public async Task<Video?> GetByIdAsync(string videoId, CancellationToken ct = default)
    {
        return await db.Videos
            .AsNoTracking()
            .FirstOrDefaultAsync(v => v.VideoId == videoId, ct);
    }

    public async Task<IReadOnlyList<Video>> GetAllAsync(CancellationToken ct = default)
    {
        return await db.Videos
            .AsNoTracking()
            .ToListAsync(ct);
    }
}
