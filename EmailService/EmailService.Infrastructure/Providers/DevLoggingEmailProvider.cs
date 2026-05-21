using EmailService.Core.Contracts;
using EmailService.Core.Models;
using Microsoft.Extensions.Logging;

namespace EmailService.Infrastructure.Providers;

/// <summary>
/// Локальная разработка без SendGrid: письмо логируется и помечается как Sent.
/// </summary>
public class DevLoggingEmailProvider : IEmailProvider
{
    private readonly IEmailStatusStore _store;
    private readonly ILogger<DevLoggingEmailProvider> _logger;

    public DevLoggingEmailProvider(IEmailStatusStore store, ILogger<DevLoggingEmailProvider> logger)
    {
        _store = store;
        _logger = logger;
    }

    public async Task SendAsync(RenderedEmail email, CancellationToken ct = default)
    {
        _logger.LogWarning(
            "DEV email (SendGrid disabled): to={To} subject={Subject} correlationId={CorrelationId}",
            email.To, email.Subject, email.CorrelationId);
        _logger.LogDebug("DEV email HTML preview ({Length} chars)", email.HtmlContent.Length);

        await _store.MarkStatusAsync(email.CorrelationId, EmailStatus.Sent, "dev-log-only", ct);
    }
}
