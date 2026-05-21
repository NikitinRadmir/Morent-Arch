using EmailService.Core.Contracts;
using EmailService.Core.Models;
using Microsoft.Extensions.Logging;

namespace EmailService.Infrastructure.Messaging;

public class MorentEmailIngestion
{
    private readonly IEmailQueue _queue;
    private readonly IEmailStatusStore _statusStore;
    private readonly ILogger<MorentEmailIngestion> _logger;

    public MorentEmailIngestion(
        IEmailQueue queue,
        IEmailStatusStore statusStore,
        ILogger<MorentEmailIngestion> logger)
    {
        _queue = queue;
        _statusStore = statusStore;
        _logger = logger;
    }

    public async Task IngestAsync(EmailRequest request, CancellationToken ct = default)
    {
        if (await _statusStore.IsProcessedAsync(request.CorrelationId, ct))
        {
            _logger.LogInformation("Email {CorrelationId} already processed, skipping kafka ingest", request.CorrelationId);
            return;
        }

        await _statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Queued, ct: ct);
        await _queue.EnqueueAsync(request, ct);
        _logger.LogInformation("Email {CorrelationId} queued from kafka (template {TemplateKey})",
            request.CorrelationId, request.TemplateKey);
    }
}
