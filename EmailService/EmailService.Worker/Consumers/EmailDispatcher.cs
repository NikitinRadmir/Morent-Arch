using EmailService.Core.Contracts;
using EmailService.Core.Exceptions;
using EmailService.Core.Models;
using Microsoft.Extensions.Logging;

namespace EmailService.Worker.Consumers;

public class EmailDispatcher
{
    private readonly ITemplateRenderer _renderer;
    private readonly IEmailProvider _provider;
    private readonly IEmailStatusStore _statusStore;
    private readonly ILogger<EmailDispatcher> _logger;

    public EmailDispatcher(
        ITemplateRenderer renderer,
        IEmailProvider provider,
        IEmailStatusStore statusStore,
        ILogger<EmailDispatcher> logger)
    {
        _renderer = renderer;
        _provider = provider;
        _statusStore = statusStore;
        _logger = logger;
    }

    public async Task ProcessAsync(EmailRequest request, CancellationToken ct = default)
    {
        if (await _statusStore.IsProcessedAsync(request.CorrelationId, ct))
        {
            _logger.LogInformation("Email {CorrelationId} already processed, skipping", request.CorrelationId);
            return;
        }

        await _statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Rendering, ct: ct);

        try
        {
            var rendered = await _renderer.RenderAsync(request, ct);
            _logger.LogDebug("Rendered template {TemplateKey} for {CorrelationId}",
                request.TemplateKey, request.CorrelationId);

            await _provider.SendAsync(rendered, ct);

        }
        catch (TemplateNotFoundException ex)
        {
            _logger.LogError(ex, "Template {TemplateKey} not found for {CorrelationId}",
                request.TemplateKey, request.CorrelationId);
            await _statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Failed,
                $"Template not found: {request.TemplateKey}", ct);
            throw; 
        }
        catch (Exception ex) when (ex is not OperationCanceledException)
        {
            _logger.LogError(ex, "Failed to process {CorrelationId}", request.CorrelationId);
            await _statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Failed,
                ex.Message, ct);
            throw;
        }
    }
}