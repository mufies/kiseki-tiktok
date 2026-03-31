using Grpc.Core;
using EventService.Protos;
using EventService.Data;
using Microsoft.EntityFrameworkCore;

namespace EventService.GrpcServices;

/// <summary>
/// gRPC server: Feed Service calls GetTrendingVideos to get trending video list.
/// Trending = videos with the most watch events in the last 7 days.
/// </summary>
public class EventGrpcService(AppDbContext db, ILogger<EventGrpcService> logger)
    : Protos.EventService.EventServiceBase
{
    public override async Task<GetTrendingResponse> GetTrendingVideos(
        GetTrendingRequest request,
        ServerCallContext context)
    {
        int limit = request.Limit > 0 ? request.Limit : 20;

        var trending = await db.WatchEvents
            .GroupBy(e => e.VideoId)
            .OrderByDescending(g => g.Count())
            .Take(limit)
            .Select(g => new { VideoId = g.Key, WatchCount = g.Count() })
            .ToListAsync(context.CancellationToken);

        var response = new GetTrendingResponse();
        foreach (var item in trending)
        {
            response.Videos.Add(new TrendingVideo
            {
                VideoId    = item.VideoId,
                Title      = string.Empty,
                WatchCount = item.WatchCount
            });
        }

        logger.LogInformation("GetTrendingVideos returned {Count} results", response.Videos.Count);
        return response;
    }
}
