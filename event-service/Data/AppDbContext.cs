using Microsoft.EntityFrameworkCore;
using EventService.Models;

namespace EventService.Data;

public class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
{
    public DbSet<WatchEvent> WatchEvents => Set<WatchEvent>();
    public DbSet<Video> Videos => Set<Video>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);

        // PostgreSQL array column for categories
        modelBuilder.Entity<Video>()
            .Property(v => v.Categories)
            .HasColumnType("text[]");

        // PostgreSQL array column for hashtags
        modelBuilder.Entity<Video>()
            .Property(v => v.Hashtags)
            .HasColumnType("text[]");

        // Index on watch_events for user queries
        modelBuilder.Entity<WatchEvent>()
            .HasIndex(e => e.UserId)
            .HasDatabaseName("ix_watch_events_user_id");

        modelBuilder.Entity<WatchEvent>()
            .HasIndex(e => new { e.UserId, e.Timestamp })
            .HasDatabaseName("ix_watch_events_user_timestamp");
    }
}
