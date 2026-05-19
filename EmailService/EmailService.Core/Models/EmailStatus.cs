namespace EmailService.Core.Models;

/// <summary>
/// Represents the processing status of an email.
/// </summary>
public enum EmailStatus
{
    /// <summary>
    /// Email request has been queued for processing.
    /// </summary>
    Queued,

    /// <summary>
    /// Email template is being rendered.
    /// </summary>
    Rendering,

    /// <summary>
    /// Email has been sent to the provider.
    /// </summary>
    Sent,

    /// <summary>
    /// Email has been successfully delivered.
    /// </summary>
    Delivered,

    /// <summary>
    /// Email delivery failed due to invalid address or mailbox issues.
    /// </summary>
    Bounced,

    /// <summary>
    /// Email sending failed after retry attempts.
    /// </summary>
    Failed,

    /// <summary>
    /// Recipient marked the email as spam.
    /// </summary>
    Spam,

    /// <summary>
    /// Recipient unsubscribed from future emails.
    /// </summary>
    Unsubscribed
}