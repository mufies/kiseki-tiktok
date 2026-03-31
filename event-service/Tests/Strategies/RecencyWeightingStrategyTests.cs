using EventService.Configuration;
using EventService.Models;
using EventService.Strategies;
using Xunit;

namespace EventService.Tests.Strategies;

public class RecencyWeightingStrategyTests
{
    private readonly WeightingConfig _defaultConfig;
    private readonly RecencyWeightingStrategy _strategy;

    public RecencyWeightingStrategyTests()
    {
        _defaultConfig = new WeightingConfig
        {
            WatchPctWeight = 0.6,
            LikeBonus = 40.0,
            LowEngagementPenalty = 20.0,
            LowEngagementThreshold = 30.0f,
            RecencyWindowDays = 30
        };
        _strategy = new RecencyWeightingStrategy(_defaultConfig);
    }

    [Fact]
    public void CalculateWeight_WithinWindow_ReturnsFullWeight()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 31);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate // Same day
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Expected: 100 * 0.6 + 40 = 100.0 (no decay on same day)
        Assert.Equal(100.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_OutsideWindow_ReturnsZero()
    {
        // Arrange
        var baseDate = new DateTime(2024, 2, 1);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-31) // 31 days old, outside 30-day window
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        Assert.Equal(0.0, weight);
    }

    [Fact]
    public void CalculateWeight_AtWindowBoundary_ReturnsZero()
    {
        // Arrange
        var baseDate = new DateTime(2024, 2, 1);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-30) // Exactly 30 days old, outside the 30-day window
        };

        // Act - Should be outside the window (window is 0-29 days)
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert - 30 days is outside the window, so weight should be 0
        Assert.Equal(0.0, weight);
    }

    [Fact]
    public void CalculateWeight_LinearDecay_CorrectWeights()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 31);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-15) // 15 days old, halfway through window
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Base weight: 100 * 0.6 + 40 = 100
        // Linear decay: 100 * (1 - 15/30) = 100 * 0.5 = 50.0
        Assert.Equal(50.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_LowEngagement_AppliesPenalty()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 31);
        var ev = new WatchEvent
        {
            WatchPct = 20.0f,
            Liked = false,
            Timestamp = baseDate
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Base weight: 20 * 0.6 - 20 = -8.0
        // No decay on same day
        Assert.Equal(-8.0, weight, precision: 2);
    }

    [Theory]
    [InlineData(0, 100.0)]     // Same day: 100 * (1 - 0/30) = 100
    [InlineData(10, 66.67)]    // 10 days: 100 * (1 - 10/30) ≈ 66.67
    [InlineData(20, 33.33)]    // 20 days: 100 * (1 - 20/30) ≈ 33.33
    [InlineData(29, 3.33)]     // 29 days: 100 * (1 - 29/30) ≈ 3.33
    public void CalculateWeight_VariousDaysInWindow_LinearDecay(int daysAgo, double expectedWeight)
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 31);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-daysAgo)
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        Assert.Equal(expectedWeight, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_CustomWindow_UsesCustomValue()
    {
        // Arrange
        var customConfig = new WeightingConfig
        {
            WatchPctWeight = 0.6,
            LikeBonus = 40.0,
            LowEngagementPenalty = 20.0,
            LowEngagementThreshold = 30.0f,
            RecencyWindowDays = 7 // Only last 7 days
        };
        var customStrategy = new RecencyWeightingStrategy(customConfig);

        var baseDate = new DateTime(2024, 1, 15);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-8) // 8 days old, outside 7-day window
        };

        // Act
        double weight = customStrategy.CalculateWeight(ev, baseDate);

        // Assert
        Assert.Equal(0.0, weight);
    }

    [Fact]
    public void CalculateWeight_NullConfig_ThrowsException()
    {
        // Assert
        Assert.Throws<ArgumentNullException>(() => new RecencyWeightingStrategy(null!));
    }
}
