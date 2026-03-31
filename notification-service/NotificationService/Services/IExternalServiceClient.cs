using NotificationService.DTOs;

namespace NotificationService.Services;

public interface IExternalServiceClient
{
    Task<UserProfileDto?> GetUserProfileAsync(string userId);
    Task<VideoInfoDto?> GetVideoInfoAsync(string videoId);
    Task<CommentInfoDto?> GetCommentInfoAsync(string commentId);
}
