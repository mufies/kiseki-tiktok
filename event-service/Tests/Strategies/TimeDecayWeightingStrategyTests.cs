using EventService.Configuration;
using EventService.Models;
using EventService.Strategies;
using Xunit;

namespace EventService.Tests.Strategies;

public class TimeDecayWeightingStrategyTests
{
    private readonly WeightingConfig _defaultConfig;
    private readonly TimeDecayWeightingStrategy _strategy;

    public TimeDecayWeightingStrategyTests()
    {
        _defaultConfig = new WeightingConfig
        {
            WatchPctWeight = 0.6,
            LikeBonus = 40.0,
            TimeDecayFactor = 0.95
        };
        _strategy = new TimeDecayWeightingStrategy(_defaultConfig);
    }

    [Fact]
    public void CalculateWeight_FullyWatchedAndLiked_NoDecay()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 1);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Expected: 100 * 0.6 + 40.0 = 100.0 (no time decay)
        Assert.Equal(100.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_NoPenaltyForLowEngagement()
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
        // TimeDecay strategy doesn't apply low engagement penalty
        // Expected: 20 * 0.6 + 0 = 12.0
        Assert.Equal(12.0, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_ExponentialDecay_MoreAggressiveThanStandard()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 10);
        var ev = new WatchEvent
        {
            WatchPct = 100.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-10)
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Base weight: 100 * 0.6 + 40 = 100
        // Lambda = -ln(0.95) ≈ 0.05129
        // Decay: e^(-0.05129 * 10) ≈ 0.5987
        // Expected: 100 * 0.5987 ≈ 59.87
        Assert.Equal(59.87, weight, precision: 2);
    }

    [Fact]
    public void CalculateWeight_RecentEvent_HighWeight()
    {
        // Arrange
        var baseDate = new DateTime(2024, 1, 5);
        var ev = new WatchEvent
        {
            WatchPct = 80.0f,
            Liked = true,
            Timestamp = baseDate.AddDays(-1)
        };

        // Act
        double weight = _strategy.CalculateWeight(ev, baseDate);

        // Assert
        // Base weight: 80 * 0.6 + 40 = 88
        // Lambda = -ln(0.95) ≈ 0.05129
        // Decay: e^(-0.05129 * 1) ≈ 0.95
        // Expected: 88 * 0.95 ≈ 83.6
        Assert.Equal(83.6, weight, precision: 2);
    }

    [Theory]
    [InlineData(0, 100.0)]    // Same day
    [InlineData(5, 77.38)]    // 5 days
    [InlineData(10, 59.87)]   // 10 days
    [InlineData(20, 35.85)]   // 20 days
    public void CalculateWeight_VariousDaysAgo_ExponentialDecay(int daysAgo, double expectedWeight)
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
    public void CalculateWeight_NullConfig_ThrowsException()
    {
        // Assert
        Assert.Throws<ArgumentNullException>(() => new TimeDecayWeightingStrategy(null!));
    }
}
