using EmailService.Core.Models;

namespace EmailService.Core.Contracts;

/// <summary>
/// Represents an email queue for storing and reading email requests.
/// </summary>
public interface IEmailQueue
{
    /// <summary>
    /// Adds an email request to the queue asynchronously.
    /// </summary>
    /// <param name="request">The email request to enqueue.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    ValueTask EnqueueAsync(EmailRequest request, CancellationToken ct = default);

    /// <summary>
    /// Reads all email requests from the queue asynchronously.
    /// </summary>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns>An asynchronous stream of email requests.</returns>
    IAsyncEnumerable<EmailRequest> ReadAllAsync(CancellationToken ct = default);
}