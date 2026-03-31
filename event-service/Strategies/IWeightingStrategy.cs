using EventService.Models;

namespace EventService.Strategies;

/// <summary>
/// Defines a strategy for calculating weight scores for watch events.
/// Implementations can provide different algorithms for scoring user engagement.
/// </summary>
public interface IWeightingStrategy
{
    /// <summary>
    /// Calculates a weight score for a watch event.
    /// </summary>
    /// <param name="ev">The watch event to score</param>
    /// <param name="baseDate">Reference date for time decay calculations (typically DateTime.UtcNow.Date)</param>
    /// <returns>A weight score representing the importance of this event</returns>
    double CalculateWeight(WatchEvent ev, DateTime baseDate);
}
