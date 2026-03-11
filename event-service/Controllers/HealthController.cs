using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using StackExchange.Redis;
using EventService.Data;

namespace EventService.Controllers;

[ApiController]
[Route("health")]
public class HealthController(AppDbContext db, IConnectionMultiplexer redis) : ControllerBase
{
    [HttpGet]
    public async Task<IActionResult> Health(CancellationToken ct)
    {
        bool pgOk    = false;
        bool redisOk = false;

        try { pgOk = await db.Database.CanConnectAsync(ct); } catch { /* ignore */ }

        try
        {
            await redis.GetDatabase().PingAsync();
            redisOk = true;
        }
        catch { /* ignore */ }

        var status = new
        {
            service    = "event-service",
            version    = "1.0",
            status     = (pgOk && redisOk) ? "healthy" : "degraded",
            postgres   = pgOk ? "ok" : "unreachable",
            redis      = redisOk ? "ok" : "unreachable"
        };

        return (pgOk && redisOk) ? Ok(status) : StatusCode(503, status);
    }
}
