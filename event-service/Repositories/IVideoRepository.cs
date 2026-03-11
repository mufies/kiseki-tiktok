using EventService.Models;

namespace EventService.Repositories;

public interface IVideoRepository
{
    Task<Video?> GetByIdAsync(string videoId, CancellationToken ct = default);
    Task<IReadOnlyList<Video>> GetAllAsync(CancellationToken ct = default);
}
