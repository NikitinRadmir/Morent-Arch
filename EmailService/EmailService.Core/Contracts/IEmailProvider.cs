using EmailService.Core.Models;

namespace EmailService.Core.Contracts;

/// <summary>
/// Provides functionality for sending emails through an external provider.
/// </summary>
public interface IEmailProvider
{
    /// <summary>
    /// Sends a rendered email message asynchronously.
    /// </summary>
    /// <param name="email">The rendered email content to send.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    Task SendAsync(RenderedEmail email, CancellationToken ct = default);
}