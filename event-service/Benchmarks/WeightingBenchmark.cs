using BenchmarkDotNet.Attributes;
using BenchmarkDotNet.Running;
using EventService.Configuration;
using EventService.Models;
using EventService.Strategies;

namespace EventService.Benchmarks;

[MemoryDiagnoser]
[SimpleJob(warmupCount: 3, iterationCount: 5)]
public class WeightingBenchmark
{
    private List<WatchEvent> _events = null!;
    private DateTime _baseDate;
    private StandardWeightingStrategy _standardStrategy = null!;
    private TimeDecayWeightingStrategy _timeDecayStrategy = null!;
    private RecencyWeightingStrategy _recencyStrategy = null!;

    [GlobalSetup]
    public void Setup()
    {
        _baseDate = DateTime.UtcNow.Date;

        // Create configuration
        var config = new WeightingConfig
        {
            WatchPctWeight = 0.6,
            LikeBonus = 40.0,
            LowEngagementPenalty = 20.0,
            LowEngagementThreshold = 30.0f,
            TimeDecayFactor = 0.95,
            RecencyWindowDays = 30
        };

        // Initialize strategies
        _standardStrategy = new StandardWeightingStrategy(config);
        _timeDecayStrategy = new TimeDecayWeightingStrategy(config);
        _recencyStrategy = new RecencyWeightingStrategy(config);

        // Generate test data: 10,000 watch events over the past 60 days
        var random = new Random(42); // Fixed seed for reproducibility
        _events = new List<WatchEvent>();

        for (int i = 0; i < 10_000; i++)
        {
            _events.Add(new WatchEvent
            {
                Id = i,
                UserId = "user123",
                VideoId = $"video{i % 1000}",
                WatchPct = (float)(random.NextDouble() * 100),
                Liked = random.NextDouble() > 0.7, // 30% like rate
                Timestamp = _baseDate.AddDays(-random.Next(0, 60))
            });
        }
    }

    [Benchmark(Baseline = true, Description = "Old Hard-coded Formula")]
    public double OldHardCodedFormula()
    {
        double totalWeight = 0;
        var now = _baseDate;

        foreach (var ev in _events)
        {
            // Original hard-coded formula
            double weight = ev.WatchPct * 0.6
                          + (ev.Liked ? 40.0 : 0.0)
                          - (ev.WatchPct < 30f ? 20.0 : 0.0);

            // Time decay
            int daysAgo = Math.Max(0, (now - ev.Timestamp.Date).Days);
            double decay = Math.Pow(0.95, daysAgo);

            totalWeight += weight * decay;
        }

        return totalWeight;
    }

    [Benchmark(Description = "Standard Strategy")]
    public double StandardStrategy()
    {
        double totalWeight = 0;

        foreach (var ev in _events)
        {
            totalWeight += _standardStrategy.CalculateWeight(ev, _baseDate);
        }

        return totalWeight;
    }

    [Benchmark(Description = "TimeDecay Strategy")]
    public double TimeDecayStrategy()
    {
        double totalWeight = 0;

        foreach (var ev in _events)
        {
            totalWeight += _timeDecayStrategy.CalculateWeight(ev, _baseDate);
        }

        return totalWeight;
    }

    [Benchmark(Description = "Recency Strategy")]
    public double RecencyStrategy()
    {
        double totalWeight = 0;

        foreach (var ev in _events)
        {
            totalWeight += _recencyStrategy.CalculateWeight(ev, _baseDate);
        }

        return totalWeight;
    }
}

public class Program
{
    public static void Main(string[] args)
    {
        Console.WriteLine("=".PadRight(80, '='));
        Console.WriteLine("EventService Weighting Strategy Performance Benchmark");
        Console.WriteLine("=".PadRight(80, '='));
        Console.WriteLine();
        Console.WriteLine("This benchmark compares the performance of:");
        Console.WriteLine("  1. Old hard-coded weighting formula (baseline)");
        Console.WriteLine("  2. New StandardWeightingStrategy (should match old behavior)");
        Console.WriteLine("  3. TimeDecayWeightingStrategy (enhanced exponential decay)");
        Console.WriteLine("  4. RecencyWeightingStrategy (sliding window approach)");
        Console.WriteLine();
        Console.WriteLine("Running benchmarks with 10,000 watch events...");
        Console.WriteLine();

        var summary = BenchmarkRunner.Run<WeightingBenchmark>();
    }
}
