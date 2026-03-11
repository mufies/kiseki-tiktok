using Microsoft.EntityFrameworkCore;
using StackExchange.Redis;
using EventService.Data;
using EventService.GrpcServices;
using EventService.Repositories;
using EventService.Services;

var builder = WebApplication.CreateBuilder(args);

// ── Controllers (REST) ────────────────────────────────────────────────────────
builder.Services.AddControllers()
    .AddJsonOptions(opts =>
        opts.JsonSerializerOptions.PropertyNamingPolicy =
            System.Text.Json.JsonNamingPolicy.CamelCase); // Use camelCase (standard for JSON)

// ── gRPC Server ───────────────────────────────────────────────────────────────
builder.Services.AddGrpc();

// ── CORS ──────────────────────────────────────────────────────────────────────
builder.Services.AddCors(o => o.AddDefaultPolicy(p =>
    p.AllowAnyOrigin().AllowAnyMethod().AllowAnyHeader()));

// ── PostgreSQL / EF Core ──────────────────────────────────────────────────────
builder.Services.AddDbContext<AppDbContext>(opts =>
    opts.UseNpgsql(builder.Configuration.GetConnectionString("PostgreSQL")));

// ── Redis ─────────────────────────────────────────────────────────────────────
builder.Services.AddSingleton<IConnectionMultiplexer>(_ =>
    ConnectionMultiplexer.Connect(builder.Configuration.GetConnectionString("Redis")!));

// ── Repositories ──────────────────────────────────────────────────────────────
builder.Services.AddScoped<IEventRepository, EventRepository>();
// VideoGrpcRepository fetches video metadata from Video Service via gRPC
builder.Services.AddScoped<IVideoRepository, VideoGrpcRepository>();

// ── Services ──────────────────────────────────────────────────────────────────
builder.Services.AddScoped<IProfileService, ProfileService>();
builder.Services.AddScoped<IWatchEventService, WatchEventService>();

var app = builder.Build();

app.UseRouting();
app.UseCors();

app.MapGrpcService<EventGrpcService>();

app.MapControllers();

// Auto-apply migrations on startup
using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    await db.Database.MigrateAsync();
}

app.Run();
