using Microsoft.AspNetCore.Mvc;
using NotificationService.Services;

namespace NotificationService.Controllers;

[ApiController]
[Route("notifications")]
public class NotificationsController : ControllerBase
{
    private readonly INotificationService _notificationService;
    private readonly ILogger<NotificationsController> _logger;

    public NotificationsController(
        INotificationService notificationService,
        ILogger<NotificationsController> logger)
    {
        _notificationService = notificationService;
        _logger = logger;
    }

    [HttpGet("{userId}/unread-count")]
    public async Task<IActionResult> GetUnreadCount(string userId)
    {
        try
        {
            var count = await _notificationService.GetUnreadCountAsync(userId);
            return Ok(new { unreadCount = count });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error getting unread count for user {UserId}", userId);
            return StatusCode(500, new { message = "Failed to get unread count" });
        }
    }

    [HttpPost("{userId}/mark-read")]
    public async Task<IActionResult> MarkAsRead(
        string userId,
        [FromBody] MarkAsReadRequest request)
    {
        try
        {
            await _notificationService.MarkAsReadAsync(userId, request.NotificationIds);
            return Ok(new { message = "Notifications marked as read" });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error marking notifications as read for user {UserId}", userId);
            return StatusCode(500, new { message = "Failed to mark notifications as read" });
        }
    }

    [HttpPost("{userId}/mark-all-read")]
    public async Task<IActionResult> MarkAllAsRead(string userId)
    {
        try
        {
            await _notificationService.MarkAllAsReadAsync(userId);
            return Ok(new { message = "All notifications marked as read" });
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error marking all notifications as read for user {UserId}", userId);
            return StatusCode(500, new { message = "Failed to mark all notifications as read" });
        }
    }

    [HttpGet("{userId}")]
    public async Task<IActionResult> GetNotifications(
        string userId,
        [FromQuery] int page = 1,
        [FromQuery] int pageSize = 20,
        [FromQuery] bool detailed = false)
    {
        try
        {
            if (detailed)
            {
                var result = await _notificationService.GetNotificationDetailsAsync(userId, page, pageSize);
                return Ok(result);
            }
            else
            {
                var result = await _notificationService.GetNotificationsAsync(userId, page, pageSize);
                return Ok(result);
            }
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error getting notifications for user {UserId}", userId);
            return StatusCode(500, new { message = "Failed to get notifications" });
        }
    }
}

public class MarkAsReadRequest
{
    public List<Guid> NotificationIds { get; set; } = new();
}
