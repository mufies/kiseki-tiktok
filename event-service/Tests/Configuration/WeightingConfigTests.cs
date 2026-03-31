using EventService.Configuration;
using Xunit;

namespace EventService.Tests.Configuration;

public class WeightingConfigTests
{
    [Fact]
    public void Validate_DefaultConfig_DoesNotThrow()
    {
        // Arrange
        var config = new WeightingConfig();

        // Act & Assert
        config.Validate(); // Should not throw
    }

    [Fact]
    public void Validate_ValidCustomConfig_DoesNotThrow()
    {
        // Arrange
        var config = new WeightingConfig
        {
            StrategyType = "TimeDecay",
            WatchPctWeight = 1.0,
            LikeBonus = 50.0,
            LowEngagementPenalty = 30.0,
            LowEngagementThreshold = 25.0f,
            TimeDecayFactor = 0.9,
            RecencyWindowDays = 14
        };

        // Act & Assert
        config.Validate(); // Should not throw
    }

    [Fact]
    public void Validate_NegativeWatchPctWeight_Throws()
    {
        // Arrange
        var config = new WeightingConfig { WatchPctWeight = -0.5 };

        // Act & Assert
        var ex = Assert.Throws<InvalidOperationException>(() => config.Validate());
        Assert.Contains("WatchPctWeight", ex.Message);
    }

    [Fact]
    public void Validate_NegativeLikeBonus_Throws()
    {
        // Arrange
        var config = new WeightingConfig { LikeBonus = -10.0 };

        // Act & Assert
        var ex = Assert.Throws<InvalidOperationException>(() => config.Validate());
        Assert.Contains("LikeBonus", ex.Message);
    }

    [Fact]
    public void Validate_NegativeLowEngagementPenalty_Throws()
    {
        // Arrange
        var config = new WeightingConfig { LowEngagementPenalty = -5.0 };

        // Act & Assert
        var ex = Assert.Throws<InvalidOperationException>(() => config.Validate());
        Assert.Contains("LowEngagementPenalty", ex.Message);
    }

    [Theory]
    [InlineData(-10.0f)]
    [InlineData(150.0f)]
    public void Validate_InvalidLowEngagementThreshold_Throws(float threshold)
    {
        // Arrange
        var config = new WeightingConfig { LowEngagementThreshold = threshold };

        // Act & Assert
        var ex = Assert.Throws<InvalidOperationException>(() => config.Validate());
        Assert.Contains("LowEngagementThreshold", ex.Message);
    }

    [Theory]
    [InlineData(0.0)]    // Zero
    [InlineData(-0.5)]   // Negative
    [InlineData(1.5)]    // Greater than 1
    public void Validate_InvalidTimeDecayFactor_Throws(double factor)
    {
        // Arrange
        var config = new WeightingConfig { TimeDecayFactor = factor };

        // Act & Assert
        var ex = Assert.Throws<InvalidOperationException>(() => config.Validate());
        Assert.Contains("TimeDecayFactor", ex.Message);
    }

    [Fact]
    public void Validate_TimeDecayFactorOne_Valid()
    {
        // Arrange
        var config = new WeightingConfig { TimeDecayFactor = 1.0 };

        // Act & Assert
        config.Validate(); // Should not throw
    }

    [Fact]
    public void Validate_ZeroRecencyWindowDays_Throws()
    {
        // Arrange
        var config = new WeightingConfig { RecencyWindowDays = 0 };

        // Act & Assert
        var ex = Assert.Throws<InvalidOperationException>(() => config.Validate());
        Assert.Contains("RecencyWindowDays", ex.Message);
    }

    [Fact]
    public void Validate_NegativeRecencyWindowDays_Throws()
    {
        // Arrange
        var config = new WeightingConfig { RecencyWindowDays = -5 };

        // Act & Assert
        var ex = Assert.Throws<InvalidOperationException>(() => config.Validate());
        Assert.Contains("RecencyWindowDays", ex.Message);
    }

    [Theory]
    [InlineData("InvalidStrategy")]
    [InlineData("")]
    [InlineData("standard")] // Case-sensitive
    public void Validate_InvalidStrategyType_Throws(string strategyType)
    {
        // Arrange
        var config = new WeightingConfig { StrategyType = strategyType };

        // Act & Assert
        var ex = Assert.Throws<InvalidOperationException>(() => config.Validate());
        Assert.Contains("StrategyType", ex.Message);
    }

    [Theory]
    [InlineData("Standard")]
    [InlineData("TimeDecay")]
    [InlineData("Recency")]
    public void Validate_ValidStrategyTypes_DoesNotThrow(string strategyType)
    {
        // Arrange
        var config = new WeightingConfig { StrategyType = strategyType };

        // Act & Assert
        config.Validate(); // Should not throw
    }
}
