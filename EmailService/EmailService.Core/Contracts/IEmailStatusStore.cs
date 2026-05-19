using EmailService.Core.Models;

namespace EmailService.Core.Contracts;

/// <summary>
/// Provides methods for storing and retrieving email processing statuses.
/// </summary>
public interface IEmailStatusStore
{
    /// <summary>
    /// Updates the status of an email request.
    /// </summary>
    /// <param name="correlationId">The unique correlation identifier of the email request.</param>
    /// <param name="status">The email processing status.</param>
    /// <param name="details">Optional additional status details.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    Task MarkStatusAsync(
        string correlationId,
        EmailStatus status,
        string? details = null,
        CancellationToken ct = default);

    /// <summary>
    /// Retrieves the status of an email request.
    /// </summary>
    /// <param name="correlationId">The unique correlation identifier of the email request.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns>The email status if found; otherwise <c>null</c>.</returns>
    Task<EmailStatus?> GetStatusAsync(
        string correlationId,
        CancellationToken ct = default);

    /// <summary>
    /// Checks whether the email request has already been processed.
    /// </summary>
    /// <param name="correlationId">The unique correlation identifier of the email request.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns><c>true</c> if the request was processed; otherwise <c>false</c>.</returns>
    Task<bool> IsProcessedAsync(
        string correlationId,
        CancellationToken ct = default);
}