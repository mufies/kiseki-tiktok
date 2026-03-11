using System.Text.Json;
using StackExchange.Redis;
using EventService.Models;
using EventService.Repositories;

namespace EventService.Services;

public class ProfileService(
    IEventRepository eventRepo,
    IVideoRepository videoRepo,
    IConnectionMultiplexer redis,
    ILogger<ProfileService> logger) : IProfileService
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        WriteIndented = false
    };

    // ─── Rebuild profile from full history and write to Redis ─────────────────
    public async Task<UserProfile> RebuildAndCacheAsync(string userId, CancellationToken ct = default)
    {
        var events = await eventRepo.GetUserEventsAsync(userId, ct);

        var categories = new Dictionary<string, double>(StringComparer.Ordinal);
        var hashtags   = new Dictionary<string, double>(StringComparer.Ordinal);
        var now        = DateTime.UtcNow.Date;

        foreach (var ev in events)
        {
            var video = await videoRepo.GetByIdAsync(ev.VideoId, ct);
            if (video is null) continue;

            // Base weight
            double weight = ev.WatchPct * 0.6
                          + (ev.Liked ? 40.0 : 0.0)
                          - (ev.WatchPct < 30f ? 20.0 : 0.0);

            // Time decay: 0.95 ^ days_ago
            int daysAgo = Math.Max(0, (now - ev.Timestamp.Date).Days);
            double decay  = Math.Pow(0.95, daysAgo);
            double finalW = weight * decay;

            foreach (var cat in video.Categories)
                categories[cat] = categories.GetValueOrDefault(cat) + finalW;

            foreach (var tag in video.Hashtags)
                hashtags[tag] = hashtags.GetValueOrDefault(tag) + finalW;
        }

        var profile = new UserProfile
        {
            UserId     = userId,
            Categories = categories,
            Hashtags   = hashtags,
            UpdatedAt  = DateTime.UtcNow
        };

        // Write to Redis with no expiry (Feed Service needs it indefinitely)
        var db  = redis.GetDatabase();
        var key = $"profile:{userId}";
        var json = JsonSerializer.Serialize(profile, JsonOptions);
        await db.StringSetAsync(key, json);

        logger.LogInformation("Rebuilt profile for user {UserId}: {Categories} categories, {Hashtags} hashtags",
            userId, categories.Count, hashtags.Count);

        return profile;
    }

    // ─── Read profile from Redis ───────────────────────────────────────────────
    public async Task<UserProfile?> GetFromCacheAsync(string userId, CancellationToken ct = default)
    {
        var db  = redis.GetDatabase();
        var key = $"profile:{userId}";
        var raw = await db.StringGetAsync(key);

        if (raw.IsNullOrEmpty) return null;

        try
        {
            return JsonSerializer.Deserialize<UserProfile>(raw.ToString(), JsonOptions);
        }
        catch (JsonException ex)
        {
            logger.LogWarning(ex, "Failed to deserialize profile for {UserId}", userId);
            return null;
        }
    }
}
