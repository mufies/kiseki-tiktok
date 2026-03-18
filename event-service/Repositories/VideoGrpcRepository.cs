using Grpc.Net.Client;
using VideoService.Protos;
using EventService.Models;

namespace EventService.Repositories;

/// <summary>
/// Fetches video metadata from Video Service via gRPC instead of hitting the DB directly.
/// </summary>
public class VideoGrpcRepository(IConfiguration config, ILogger<VideoGrpcRepository> logger) : IVideoRepository
{
    // Channel is thread-safe and long-lived; one per process.
    private readonly VideoService.Protos.VideoService.VideoServiceClient _client = new(
        GrpcChannel.ForAddress(config["VideoService:GrpcAddress"] ?? "http://localhost:9091"));

    public async Task<Models.Video?> GetByIdAsync(string videoId, CancellationToken ct = default)
    {
        try
        {
            var response = await _client.GetVideoAsync(new VideoService.Protos.GetVideoRequest { VideoId = videoId },
                cancellationToken: ct);

            return MapProtoToModel(response.Video);
        }
        catch (Exception ex)
        {
            logger.LogWarning(ex, "gRPC GetVideo failed for {VideoId}", videoId);
            return null;
        }
    }

    public async Task<IReadOnlyList<Models.Video>> GetAllAsync(CancellationToken ct = default)
    {
        try
        {
            var response = await _client.GetVideosAsync(new VideoService.Protos.GetVideosRequest { Limit = 1000, Offset = 0 },
                cancellationToken: ct);

            return response.Videos.Select(MapProtoToModel).ToList();
        }
        catch (Exception ex)
        {
            logger.LogWarning(ex, "gRPC GetVideos failed");
            return [];
        }
    }

    private static Models.Video MapProtoToModel(VideoService.Protos.Video v) => new()
    {
        VideoId    = v.VideoId,
        Title      = v.Title,
        Categories = [.. v.Categories],
        Hashtags   = [.. v.Hashtags]
    };
}
