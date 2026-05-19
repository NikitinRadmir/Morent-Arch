using EmailService.Core.Contracts;
using EmailService.Core.Exceptions;
using EmailService.Core.Models;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;

namespace EmailService.Worker.Consumers;

public class EmailDispatcher
{
    private readonly ITemplateRenderer _renderer;
    private readonly IEmailProvider _provider;
    private readonly IServiceScopeFactory _scopeFactory; // ← Вместо прямого IEmailStatusStore
    private readonly ILogger<EmailDispatcher> _logger;

    public EmailDispatcher(
        ITemplateRenderer renderer,
        IEmailProvider provider,
        IServiceScopeFactory scopeFactory, // ← Inject scope factory
        ILogger<EmailDispatcher> logger)
    {
        _renderer = renderer;
        _provider = provider;
        _scopeFactory = scopeFactory;
        _logger = logger;
    }

    public async Task ProcessAsync(EmailRequest request, CancellationToken ct = default)
    {
        // 🔹 Создаём scope для каждого сообщения — внутри него можно резолвить Scoped сервисы
        using var scope = _scopeFactory.CreateScope();
        var statusStore = scope.ServiceProvider.GetRequiredService<IEmailStatusStore>();

        try
        {
            // 1. Idempotency check
            if (await statusStore.IsProcessedAsync(request.CorrelationId, ct))
            {
                _logger.LogInformation("Email {CorrelationId} already processed, skipping", request.CorrelationId);
                return;
            }

            // 2. Статус "Rendering"
            await statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Rendering, ct: ct);

            // 3. Рендеринг шаблона
            var rendered = await _renderer.RenderAsync(request, ct);
            _logger.LogDebug("Rendered template {TemplateKey} for {CorrelationId}",
                request.TemplateKey, request.CorrelationId);

            // 4. Отправка (Polly retry уже внутри провайдера, если настроен)
            await _provider.SendAsync(rendered, ct);

            // Статус "Sent" уже проставлен внутри провайдера при успехе
        }
        catch (TemplateNotFoundException ex)
        {
            _logger.LogError(ex, "Template {TemplateKey} not found for {CorrelationId}",
                request.TemplateKey, request.CorrelationId);
            await statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Failed,
                $"Template not found: {request.TemplateKey}", ct);
            throw;
        }
        catch (Exception ex) when (ex is not OperationCanceledException)
        {
            _logger.LogError(ex, "Failed to process {CorrelationId}", request.CorrelationId);
            await statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Failed,
                ex.Message, ct);
            throw;
        }
    }
}