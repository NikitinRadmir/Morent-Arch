namespace EmailService.Core.Models;

/// <summary>
/// Represents a request for sending an email.
/// </summary>
public record EmailRequest
{
    /// <summary>
    /// Unique identifier used to track the email request.
    /// </summary>
    public string CorrelationId { get; init; } = Guid.NewGuid().ToString("N");

    /// <summary>
    /// Recipient email address.
    /// </summary>
    public string To { get; init; } = null!;

    /// <summary>
    /// Template key used for email rendering.
    /// </summary>
    public string TemplateKey { get; init; } = null!;

    /// <summary>
    /// Template variables used during rendering.
    /// </summary>
    public IReadOnlyDictionary<string, string> Variables { get; init; }
        = new Dictionary<string, string>();

    /// <summary>
    /// Priority of the email request.
    /// </summary>
    public EmailPriority Priority { get; init; } = EmailPriority.Normal;
}

/// <summary>
/// Defines email processing priority levels.
/// </summary>
public enum EmailPriority
{
    /// <summary>
    /// Low priority email.
    /// </summary>
    Low = 0,

    /// <summary>
    /// Normal priority email.
    /// </summary>
    Normal = 1,

    /// <summary>
    /// High priority email.
    /// </summary>
    High = 2
}