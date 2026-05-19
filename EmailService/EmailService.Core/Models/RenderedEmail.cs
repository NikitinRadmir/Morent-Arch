namespace EmailService.Core.Models;

/// <summary>
/// Represents a fully rendered email ready for delivery.
/// </summary>
public record RenderedEmail
{
    /// <summary>
    /// Unique identifier used to track the email request.
    /// </summary>
    public string CorrelationId { get; init; } = null!;

    /// <summary>
    /// Recipient email address.
    /// </summary>
    public string To { get; init; } = null!;

    /// <summary>
    /// Subject of the email.
    /// </summary>
    public string Subject { get; init; } = null!;

    /// <summary>
    /// HTML content of the email.
    /// </summary>
    public string HtmlContent { get; init; } = null!;

    /// <summary>
    /// Priority of the email.
    /// </summary>
    public EmailPriority Priority { get; init; } = EmailPriority.Normal;
}