using System.Text.Json;
using Microsoft.Extensions.Options;
using NotificationService.Configuration;
using NotificationService.DTOs;

namespace NotificationService.Services;

public class ExternalServiceClient : IExternalServiceClient
{
    private readonly HttpClient _httpClient;
    private readonly ExternalServiceSettings _settings;
    private readonly ILogger<ExternalServiceClient> _logger;

    public ExternalServiceClient(
        HttpClient httpClient,
        IOptions<ExternalServiceSettings> settings,
        ILogger<ExternalServiceClient> logger)
    {
        _httpClient = httpClient;
        _settings = settings.Value;
        _logger = logger;
    }

    public async Task<UserProfileDto?> GetUserProfileAsync(string userId)
    {
        try
        {
            var url = $"{_settings.UserServiceUrl}/users/{userId}";
            var response = await _httpClient.GetAsync(url);

            if (!response.IsSuccessStatusCode)
            {
                _logger.LogWarning("Failed to get user profile for {UserId}: {StatusCode}", userId, response.StatusCode);
                return null;
            }

            var content = await response.Content.ReadAsStringAsync();
            var user = JsonSerializer.Deserialize<UserProfileDto>(content, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });

            return user;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error fetching user profile for {UserId}", userId);
            return null;
        }
    }

    public async Task<VideoInfoDto?> GetVideoInfoAsync(string videoId)
    {
        try
        {
            var url = $"{_settings.VideoServiceUrl}/videos/{videoId}";
            var response = await _httpClient.GetAsync(url);

            if (!response.IsSuccessStatusCode)
            {
                _logger.LogWarning("Failed to get video info for {VideoId}: {StatusCode}", videoId, response.StatusCode);
                return null;
            }

            var content = await response.Content.ReadAsStringAsync();
            var video = JsonSerializer.Deserialize<VideoInfoDto>(content, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });

            return video;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error fetching video info for {VideoId}", videoId);
            return null;
        }
    }

    public async Task<CommentInfoDto?> GetCommentInfoAsync(string commentId)
    {
        try
        {
            var url = $"{_settings.InteractionServiceUrl}/api/comments/{commentId}";
            var response = await _httpClient.GetAsync(url);

            if (!response.IsSuccessStatusCode)
            {
                _logger.LogWarning("Failed to get comment info for {CommentId}: {StatusCode}", commentId, response.StatusCode);
                return null;
            }

            var content = await response.Content.ReadAsStringAsync();
            var comment = JsonSerializer.Deserialize<CommentInfoDto>(content, new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            });

            return comment;
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Error fetching comment info for {CommentId}", commentId);
            return null;
        }
    }
}
