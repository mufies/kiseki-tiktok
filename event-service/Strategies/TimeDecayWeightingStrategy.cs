using EventService.Configuration;
using EventService.Models;

namespace EventService.Strategies;

/// <summary>
/// Enhanced time decay strategy with exponential decay.
/// Places more emphasis on recent events with configurable decay rate.
/// Useful for rapidly changing user preferences.
/// </summary>
public class TimeDecayWeightingStrategy : IWeightingStrategy
{
    private readonly WeightingConfig _config;

    public TimeDecayWeightingStrategy(WeightingConfig config)
    {
        _config = config ?? throw new ArgumentNullException(nameof(config));
    }

    public double CalculateWeight(WatchEvent ev, DateTime baseDate)
    {
        double baseWeight = ev.WatchPct * _config.WatchPctWeight
                          + (ev.Liked ? _config.LikeBonus : 0.0);

        int daysAgo = Math.Max(0, (baseDate - ev.Timestamp.Date).Days);

        // e^(-lambda * days) where lambda = -ln(decayFactor)
        double lambda = -Math.Log(_config.TimeDecayFactor);
        double decay = Math.Exp(-lambda * daysAgo);

        return baseWeight * decay;
    }
}
