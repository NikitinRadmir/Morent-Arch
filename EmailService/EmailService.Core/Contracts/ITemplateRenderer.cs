using EmailService.Core.Models;

namespace EmailService.Core.Contracts;

/// <summary>
/// Provides functionality for rendering email templates.
/// </summary>
public interface ITemplateRenderer
{
    /// <summary>
    /// Renders an email template based on the provided request.
    /// </summary>
    /// <param name="request">The email request containing template data.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns>A rendered email instance.</returns>
    Task<RenderedEmail> RenderAsync(
        EmailRequest request,
        CancellationToken ct = default);
}