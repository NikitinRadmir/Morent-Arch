using EmailService.Core.Contracts;
using EmailService.Core.Models;
using MassTransit;
using Microsoft.Extensions.Logging;

namespace EmailService.Infrastructure.Consumers;

/// <summary>
/// Consumes email requests from the message broker and processes them.
/// </summary>
public class EmailConsumer : IConsumer<EmailRequest>
{
    private readonly ITemplateRenderer _renderer;
    private readonly IEmailProvider _provider;
    private readonly IEmailStatusStore _statusStore;
    private readonly ILogger<EmailConsumer> _logger;

    /// <summary>
    /// Initializes a new instance of the <see cref="EmailConsumer"/> class.
    /// </summary>
    /// <param name="renderer">Email template renderer.</param>
    /// <param name="provider">Email delivery provider.</param>
    /// <param name="statusStore">Email status storage service.</param>
    /// <param name="logger">Logger instance.</param>
    public EmailConsumer(
        ITemplateRenderer renderer,
        IEmailProvider provider,
        IEmailStatusStore statusStore,
        ILogger<EmailConsumer> logger)
    {
        _renderer = renderer;
        _provider = provider;
        _statusStore = statusStore;
        _logger = logger;
    }

    /// <summary>
    /// Consumes and processes an email request message.
    /// </summary>
    /// <param name="context">MassTransit consume context.</param>
    public async Task Consume(ConsumeContext<EmailRequest> context)
    {
        var message = context.Message;

        using (Serilog.Context.LogContext.PushProperty("CorrelationId", message.CorrelationId))
        {
            _logger.LogInformation(
                "Processing email {CorrelationId} from RabbitMQ: Template={TemplateKey}, To={To}",
                message.CorrelationId,
                message.TemplateKey,
                message.To);

            try
            {
                await _statusStore.MarkStatusAsync(
                    message.CorrelationId,
                    EmailStatus.Rendering);

                _logger.LogDebug("Status marked as Rendering");

                var rendered = await _renderer.RenderAsync(message);

                _logger.LogDebug("Template rendered successfully");

                await _provider.SendAsync(rendered);

                _logger.LogDebug("Email sent via provider");

                await _statusStore.MarkStatusAsync(
                    message.CorrelationId,
                    EmailStatus.Sent);

                _logger.LogInformation(
                    "Email {CorrelationId} sent successfully",
                    message.CorrelationId);
            }
            catch (Exception ex) when (ex is not OperationCanceledException)
            {
                _logger.LogError(
                    ex,
                    "Failed to process email {CorrelationId}",
                    message.CorrelationId);

                await _statusStore.MarkStatusAsync(
                    message.CorrelationId,
                    EmailStatus.Failed,
                    ex.Message);

                throw;
            }
        }
    }
}