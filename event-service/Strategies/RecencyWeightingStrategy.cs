using EventService.Configuration;
using EventService.Models;

namespace EventService.Strategies;

/// <summary>
/// Recency-focused weighting strategy with sliding window.
/// Completely ignores events older than a threshold (default 30 days).
/// Provides linear decay within the window for maximum recency emphasis.
/// </summary>
public class RecencyWeightingStrategy : IWeightingStrategy
{
    private readonly WeightingConfig _config;

    public RecencyWeightingStrategy(WeightingConfig config)
    {
        _config = config ?? throw new ArgumentNullException(nameof(config));
    }

    public double CalculateWeight(WatchEvent ev, DateTime baseDate)
    {
        int daysAgo = Math.Max(0, (baseDate - ev.Timestamp.Date).Days);

        if (daysAgo > _config.RecencyWindowDays)
            return 0.0;

        double baseWeight = ev.WatchPct * _config.WatchPctWeight
                          + (ev.Liked ? _config.LikeBonus : 0.0)
                          - (ev.WatchPct < _config.LowEngagementThreshold ? _config.LowEngagementPenalty : 0.0);

        double decayFactor = 1.0 - (double)daysAgo / _config.RecencyWindowDays;

        return baseWeight * decayFactor;
    }
}
