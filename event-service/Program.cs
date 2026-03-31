using Microsoft.EntityFrameworkCore;
using StackExchange.Redis;
using EventService.Configuration;
using EventService.Data;
using EventService.GrpcServices;
using EventService.Repositories;
using EventService.Services;
using EventService.Strategies;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddControllers()
    .AddJsonOptions(opts =>
        opts.JsonSerializerOptions.PropertyNamingPolicy =
            System.Text.Json.JsonNamingPolicy.CamelCase);

builder.Services.AddGrpc();

builder.Services.AddCors(o => o.AddDefaultPolicy(p =>
    p.AllowAnyOrigin().AllowAnyMethod().AllowAnyHeader()));

builder.Services.AddDbContext<AppDbContext>(opts =>
    opts.UseNpgsql(builder.Configuration.GetConnectionString("PostgreSQL")));

builder.Services.AddSingleton<IConnectionMultiplexer>(_ =>
    ConnectionMultiplexer.Connect(builder.Configuration.GetConnectionString("Redis")!));

var weightingConfig = builder.Configuration
    .GetSection("Weighting")
    .Get<WeightingConfig>() ?? new WeightingConfig();

weightingConfig.Validate();
builder.Services.AddSingleton(weightingConfig);

builder.Services.AddScoped<IWeightingStrategy>(sp =>
{
    var config = sp.GetRequiredService<WeightingConfig>();
    return config.StrategyType switch
    {
        "Standard" => new StandardWeightingStrategy(config),
        "TimeDecay" => new TimeDecayWeightingStrategy(config),
        "Recency" => new RecencyWeightingStrategy(config),
        _ => throw new InvalidOperationException($"Unknown strategy type: {config.StrategyType}")
    };
});

builder.Services.AddScoped<IEventRepository, EventRepository>();
builder.Services.AddScoped<IVideoRepository, VideoGrpcRepository>();

builder.Services.AddScoped<IProfileService, ProfileService>();
builder.Services.AddScoped<IWatchEventService, WatchEventService>();

var app = builder.Build();

app.UseRouting();
app.UseCors();

app.MapGrpcService<EventGrpcService>();

app.MapControllers();

using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    await db.Database.MigrateAsync();
}

app.Run();
