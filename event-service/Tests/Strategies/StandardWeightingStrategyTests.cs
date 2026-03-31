using EventService.Configuration;
using EventService.Models;
using EventService.Strategies;
using Xunit;

namespace EventService.Tests.Strategies;

public class StandardWeightingStrategyTests
{
    private readonly WeightingConfig _defaultConfig;
    private readonly StandardWeightingStrategy _strategy;

    public StandardWeightingStrategyTests()
    {
        _defaultConfig = new WeightingConfig
        {
            WatchPctWeight = 0.6,
            LikeBonus = 40.0,
            LowEngagementPenalty = 20.0,
            LowEngagementThreshold = 30.0f,
            TimeDecayFactor = 0.95
        };
        _strategy = new StandardWeightingStrategy(_defaultConfig);
    }

    [Fact]
    public void CalculateWeight_FullyWatchedAndLiked_ReturnsMaxWeight()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 1);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate // Same day, no decay
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Expected: 100 * 0.6 + 40.0 = 100.0 (no penalty as WatchPct >= 30)
        Assert.Equal(100.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_LowEngagement_AppliesPenalty()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 1);
        var ev = new WatchEvent
        {
            WatchPct = 20.0f,
            Liked = false,
            Timestamp = baseDate
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Expected: 20 * 0.6 + 0 - 20.0 = -8.0
        Assert.Equal(-8.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_NoPenaltyAtThreshold()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 1);
        var ev = new WatchEvent
        {
            WatchPct = 30.0f,
            Liked = false,
            Timestamp = baseDate
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Expected: 30 * 0.6 + 0 - 0 = 18.0 (at threshold, no penalty)
        Assert.Equal(18.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_TimeDecay_AppliesCorrectly()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 10);
        var ev = new WatchEvent
        {
            WatchPct = 50.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-5) // 5 days ago
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Base weight: 50 * 0.6 + 40 = 70
        // Decay: 0.95^5 ≈ 0.7738
        // Expected: 70 * 0.7738 ≈ 54.16
        Assert.Equal(54.16, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_OneDayOld_AppliesDecay()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 2);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-1)
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Base weight: 100 * 0.6 + 40 = 100
        // Decay: 0.95^1 = 0.95
        // Expected: 100 * 0.95 = 95.0
        Assert.Equal(95.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_CustomConfig_UsesCustomValues()
    {
        // Arrange
        var customConfig = new WeightingConfig
        {
            WatchPctWeight = 1.0,
            LikeBonus = 50.0,
            LowEngagementPenalty = 30.0,
            LowEngagementThreshold = 50.0f,
            TimeDecayFactor = 0.9
        };
        var customStrategy = new StandardWeightingStrategy(customConfig);

        var baseDate = new DateTime(2024, 1, 1);
        var ev = new WatchEvent
        {
            WatchPct = 60.0f,
            Liked = true,
            Timestamp = baseDate
        };

        // Act
        double weight = customStrategy.CalculateWeight(ev, baseDate);

        // Assert
        // Expected: 60 * 1.0 + 50.0 = 110.0 (no penalty as WatchPct >= 50)
        Assert.Equal(110.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_NullConfig_ThrowsException()
    {
        // Assert
        Assert.Throws<ArgumentNullException>(() => new StandardWeightingStrategy(null!));
    }

    [Theory]
    [InlineData(0, 100.0)]    // Same day
    [InlineData(10, 59.87)]   // 10 days: 100 * 0.95^10
    [InlineData(30, 21.46)]   // 30 days: 100 * 0.95^30
    public void CalculateWeight_VariousDaysAgo_CorrectDecay(int daysAgo, double expectedWeight)
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
}
