using MailKit.Net.Smtp;
using MimeKit;
using NotificationService.Configuration;
using Microsoft.Extensions.Options;

namespace NotificationService.Services;

public class EmailService : IEmailService
{
    private readonly SmtpSettings _smtpSettings;
    private readonly ILogger<EmailService> _logger;

    public EmailService(
        IOptions<SmtpSettings> smtpSettings,
        ILogger<EmailService> logger)
    {
        _smtpSettings = smtpSettings.Value;
        _logger = logger;
    }

    public async Task SendFollowNotificationAsync(string toEmail, string fromUsername)
    {
        try
        {
            var message = new MimeMessage();
            message.From.Add(new MailboxAddress(_smtpSettings.FromName, _smtpSettings.FromEmail));
            message.To.Add(new MailboxAddress("", toEmail));
            message.Subject = $"{fromUsername} started following you!";

            var bodyBuilder = new BodyBuilder
            {
                HtmlBody = $@"
                    <html>
                    <body style='font-family: Arial, sans-serif;'>
                        <div style='max-width: 600px; margin: 0 auto; padding: 20px;'>
                            <h2 style='color: #fe2c55;'>New Follower!</h2>
                            <p>Hi there!</p>
                            <p><strong>{fromUsername}</strong> just started following you on TikTok Clone.</p>
                            <p>Check out their profile and see what they're sharing!</p>
                            <div style='margin-top: 30px; padding: 20px; background-color: #f8f8f8; border-radius: 8px;'>
                                <a href='#' style='display: inline-block; padding: 12px 24px; background-color: #fe2c55; color: white; text-decoration: none; border-radius: 4px;'>
                                    View Profile
                                </a>
                            </div>
                            <p style='margin-top: 30px; color: #888; font-size: 12px;'>
                                This is an automated notification from TikTok Clone.
                            </p>
                        </div>
                    </body>
                    </html>
                "
            };

            message.Body = bodyBuilder.ToMessageBody();

            using var client = new SmtpClient();
            await client.ConnectAsync(_smtpSettings.Host, _smtpSettings.Port, MailKit.Security.SecureSocketOptions.StartTls);
            await client.AuthenticateAsync(_smtpSettings.Username, _smtpSettings.Password);
            await client.SendAsync(message);
            await client.DisconnectAsync(true);

            _logger.LogInformation("Follow notification email sent to {Email}", toEmail);
        }
        catch (Exception ex)
        {
            _logger.LogError(ex, "Failed to send follow notification email to {Email}", toEmail);
        }
    }
}
