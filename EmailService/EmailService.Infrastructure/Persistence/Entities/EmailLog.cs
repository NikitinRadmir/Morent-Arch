using EmailService.Core.Models;

namespace EmailService.Infrastructure.Persistence.Entities;

/// <summary>
/// Represents an email processing log entry.
/// </summary>
public class EmailLog
{
    /// <summary>
    /// Unique identifier used to track the email request.
    /// </summary>
    public string CorrelationId { get; set; } = null!;

    /// <summary>
    /// Current email processing status.
    /// </summary>
    public EmailStatus Status { get; set; }

    /// <summary>
    /// Identifier returned by the email provider.
    /// </summary>
    public string? ProviderMessageId { get; set; }

    /// <summary>
    /// Number of processing or delivery attempts.
    /// </summary>
    public int Attempts { get; set; }

    /// <summary>
    /// Error details related to email processing or delivery.
    /// </summary>
    public string? ErrorDetails { get; set; }

    /// <summary>
    /// Date and time when the log entry was created.
    /// </summary>
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

    /// <summary>
    /// Date and time when the log entry was last updated.
    /// </summary>
    public DateTime? UpdatedAt { get; set; }
}