namespace NotificationService.Services;

public interface IEmailService
{
    Task SendFollowNotificationAsync(string toEmail, string fromUsername);
}
