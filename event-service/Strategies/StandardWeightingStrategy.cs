using EventService.Configuration;
using EventService.Models;

namespace EventService.Strategies;

/// <summary>
/// Standard weighting strategy that maintains the original algorithm.
/// Combines watch percentage, like bonus, low engagement penalty, and time decay.
/// </summary>
public class StandardWeightingStrategy : IWeightingStrategy
{
    private readonly WeightingConfig _config;

    public StandardWeightingStrategy(WeightingConfig config)
    {
        _config = config ?? throw new ArgumentNullException(nameof(config));
    }

    public double CalculateWeight(WatchEvent ev, DateTime baseDate)
    {
        double baseWeight = ev.WatchPct * _config.WatchPctWeight
                          + (ev.Liked ? _config.LikeBonus : 0.0)
                          - (ev.WatchPct < _config.LowEngagementThreshold ? _config.LowEngagementPenalty : 0.0);

        int daysAgo = Math.Max(0, (baseDate - ev.Timestamp.Date).Days);
        double decay = Math.Pow(_config.TimeDecayFactor, daysAgo);

        return baseWeight * decay;
    }
}
