using Grpc.Core;
using NotificationService.Grpc;
using NotificationService.Services;

namespace NotificationService.GrpcServices;

public class NotificationGrpcService : Grpc.NotificationService.NotificationServiceBase
{
    private readonly INotificationService _notificationService;
    private readonly ILogger<NotificationGrpcService> _logger;

    public NotificationGrpcService(
        INotificationService notificationService,
        ILogger<NotificationGrpcService> logger)
    {
        _notificationService = notificationService;
        _logger = logger;
    }

    public override async Task<GetNotificationsResponse> GetNotifications(
        GetNotificationsRequest request,
        ServerCallContext context)
    {
        try
        {
            // Input validation
            if (string.IsNullOrWhiteSpace(request.UserId))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "UserId is required"));
            }

            // Validate pagination parameters
            if (request.Page < 1 || request.Page > 1000)
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "Page must be between 1 and 1000"));
            }

            if (request.PageSize < 1 || request.PageSize > 100)
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "PageSize must be between 1 and 100"));
            }

            // TODO: Add authorization check - verify that the authenticated user matches request.UserId
            // Example: var authenticatedUserId = context.GetHttpContext().User.FindFirst("sub")?.Value;
            // if (authenticatedUserId != request.UserId) throw new RpcException(...)

            var result = await _notificationService.GetNotificationsAsync(
                request.UserId,
                request.Page,
                request.PageSize);

            var response = new GetNotificationsResponse
            {
                TotalCount = result.TotalCount,
                Page = result.Page,
                PageSize = result.PageSize,
                TotalPages = result.TotalPages
            };

            foreach (var notification in result.Items)
            {
                var grpcNotification = new Notification
                {
                    Id = notification.Id.ToString(),
                    UserId = notification.UserId,
                    FromUserId = notification.FromUserId,
                    Type = notification.Type switch
                    {
                        Models.NotificationType.Like => NotificationType.Like,
                        Models.NotificationType.Comment => NotificationType.Comment,
                        Models.NotificationType.Follow => NotificationType.Follow,
                        Models.NotificationType.Bookmark => NotificationType.Bookmark,
                        _ => NotificationType.Like
                    },
                    IsRead = notification.IsRead,
                    CreatedAt = notification.CreatedAt.Ticks
                };

                if (notification.VideoId != null)
                    grpcNotification.VideoId = notification.VideoId;

                if (notification.CommentId != null)
                    grpcNotification.CommentId = notification.CommentId;

                response.Notifications.Add(grpcNotification);
            }

            return response;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error getting notifications for user {UserId}", request.UserId);
            throw new RpcException(new Status(StatusCode.Internal, "Failed to get notifications"));
        }
    }

    public override async Task<MarkAsReadResponse> MarkAsRead(
        MarkAsReadRequest request,
        ServerCallContext context)
    {
        try
        {
            // Input validation
            if (string.IsNullOrWhiteSpace(request.UserId))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "UserId is required"));
            }

            if (request.NotificationIds == null || request.NotificationIds.Count == 0)
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "NotificationIds cannot be empty"));
            }

            // Prevent abuse - limit batch size
            if (request.NotificationIds.Count > 100)
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "Cannot mark more than 100 notifications at once"));
            }

            // TODO: Add authorization check - verify that the authenticated user matches request.UserId
            // Example: var authenticatedUserId = context.GetHttpContext().User.FindFirst("sub")?.Value;
            // if (authenticatedUserId != request.UserId) throw new RpcException(...)

            List<Guid> notificationIds;
            try
            {
                notificationIds = request.NotificationIds
                    .Select(id => Guid.Parse(id))
                    .ToList();
            }
            catch (FormatException)
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "Invalid notification ID format"));
            }

            await _notificationService.MarkAsReadAsync(request.UserId, notificationIds);

            return new MarkAsReadResponse
            {
                Success = true,
                Message = "Notifications marked as read"
            };
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error marking notifications as read for user {UserId}", request.UserId);
            return new MarkAsReadResponse
            {
                Success = false,
                Message = "Failed to mark notifications as read"
            };
        }
    }

    public override async Task<GetUnreadCountResponse> GetUnreadCount(
        GetUnreadCountRequest request,
        ServerCallContext context)
    {
        try
        {
            // Input validation
            if (string.IsNullOrWhiteSpace(request.UserId))
            {
                throw new RpcException(new Status(StatusCode.InvalidArgument, "UserId is required"));
            }

            // TODO: Add authorization check - verify that the authenticated user matches request.UserId
            // Example: var authenticatedUserId = context.GetHttpContext().User.FindFirst("sub")?.Value;
            // if (authenticatedUserId != request.UserId) throw new RpcException(...)

            var count = await _notificationService.GetUnreadCountAsync(request.UserId);

            return new GetUnreadCountResponse
            {
                UnreadCount = count
            };
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error getting unread count for user {UserId}", request.UserId);
            throw new RpcException(new Status(StatusCode.Internal, "Failed to get unread count"));
        }
    }
}
