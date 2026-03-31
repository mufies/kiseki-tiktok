namespace EventService.Configuration;

/// <summary>
/// Configuration for weighting strategies.
/// Values can be overridden in appsettings.json under "Weighting" section.
/// </summary>
public class WeightingConfig
{
    /// <summary>
    /// The strategy type to use for weight calculation.
    /// Options: "Standard", "TimeDecay", "Recency"
    /// </summary>
    public string StrategyType { get; set; } = "Standard";

    /// <summary>
    /// Multiplier for watch percentage (0-100 scale).
    /// Default: 0.6 (matches original implementation)
    /// </summary>
    public double WatchPctWeight { get; set; } = 0.6;

    /// <summary>
    /// Bonus points added when user liked the video.
    /// Default: 40.0 (matches original implementation)
    /// </summary>
    public double LikeBonus { get; set; } = 40.0;

    /// <summary>
    /// Penalty for videos watched below the threshold percentage.
    /// Default: 20.0 (matches original implementation)
    /// </summary>
    public double LowEngagementPenalty { get; set; } = 20.0;

    /// <summary>
    /// Watch percentage threshold for low engagement penalty.
    /// Default: 30.0 (matches original implementation)
    /// </summary>
    public float LowEngagementThreshold { get; set; } = 30.0f;

    /// <summary>
    /// Time decay factor applied per day (exponential decay base).
    /// Default: 0.95 means 5% decay per day (matches original implementation)
    /// </summary>
    public double TimeDecayFactor { get; set; } = 0.95;

    /// <summary>
    /// Number of days to consider for recency-based strategies.
    /// Events older than this are ignored when using RecencyWeightingStrategy.
    /// Default: 30 days
    /// </summary>
    public int RecencyWindowDays { get; set; } = 30;

    /// <summary>
    /// Validates the configuration values.
    /// </summary>
    public void Validate()
    {
        if (WatchPctWeight < 0)
            throw new InvalidOperationException("WatchPctWeight must be non-negative");

        if (LikeBonus < 0)
            throw new InvalidOperationException("LikeBonus must be non-negative");

        if (LowEngagementPenalty < 0)
            throw new InvalidOperationException("LowEngagementPenalty must be non-negative");

        if (LowEngagementThreshold < 0 || LowEngagementThreshold > 100)
            throw new InvalidOperationException("LowEngagementThreshold must be between 0 and 100");

        if (TimeDecayFactor <= 0 || TimeDecayFactor > 1)
            throw new InvalidOperationException("TimeDecayFactor must be between 0 (exclusive) and 1 (inclusive)");

        if (RecencyWindowDays < 1)
            throw new InvalidOperationException("RecencyWindowDays must be at least 1");

        if (!new[] { "Standard", "TimeDecay", "Recency" }.Contains(StrategyType))
            throw new InvalidOperationException($"Invalid StrategyType: {StrategyType}. Must be Standard, TimeDecay, or Recency");
    }
}
