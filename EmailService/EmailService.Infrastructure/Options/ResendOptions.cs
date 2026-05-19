namespace EmailService.Infrastructure.Options;

/// <summary>
/// Configuration options for Resend email provider.
/// </summary>
public class ResendOptions
{
    /// <summary>
    /// API key used for authentication with Resend.
    /// </summary>
    public string ApiKey { get; set; } = null!;

    /// <summary>
    /// Sender email address.
    /// </summary>
    public string FromEmail { get; set; } = null!;

    /// <summary>
    /// Sender display name.
    /// </summary>
    public string FromName { get; set; } = "AutoRental";

    /// <summary>
    /// Optional webhook secret used for webhook validation.
    /// </summary>
    public string? WebhookSecret { get; set; }
}