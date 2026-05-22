using Confluent.Kafka;
using EmailService.Core.Contracts;
using EmailService.Core.Models;
using EmailService.Infrastructure.Options;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace EmailService.Infrastructure.Messaging;

/// <summary>
/// Читает <c>morent.email.send</c> из Kafka и отправляет письма через SMTP/провайдер.
/// </summary>
public sealed class KafkaEmailConsumer : BackgroundService
{
    private readonly KafkaOptions _options;
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly ILogger<KafkaEmailConsumer> _logger;
    private IConsumer<string, byte[]>? _consumer;

    public KafkaEmailConsumer(
        IOptions<KafkaOptions> options,
        IServiceScopeFactory scopeFactory,
        ILogger<KafkaEmailConsumer> logger)
    {
        _options = options.Value;
        _scopeFactory = scopeFactory;
        _logger = logger;
    }

    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        if (!_options.Enabled)
        {
            _logger.LogInformation("Kafka consumer disabled (Kafka:Enabled=false)");
            return;
        }

        var brokers = _options.Brokers
            .Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);

        if (brokers.Length == 0)
        {
            _logger.LogError("Kafka brokers are not configured");
            return;
        }

        var config = new ConsumerConfig
        {
            BootstrapServers = string.Join(",", brokers),
            GroupId = _options.GroupId,
            AutoOffsetReset = AutoOffsetReset.Earliest,
            EnableAutoCommit = false,
            AllowAutoCreateTopics = true
        };

        _consumer = new ConsumerBuilder<string, byte[]>(config).Build();
        _consumer.Subscribe(_options.Topic);

        _logger.LogInformation(
            "Kafka email consumer started: topic={Topic}, group={GroupId}, brokers={Brokers}",
            _options.Topic,
            _options.GroupId,
            _options.Brokers);

        try
        {
            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    var result = _consumer.Consume(stoppingToken);
                    await HandleMessageAsync(result, stoppingToken);
                }
                catch (ConsumeException ex)
                {
                    _logger.LogError(ex, "Kafka consume error");
                    await Task.Delay(TimeSpan.FromSeconds(1), stoppingToken);
                }
            }
        }
        catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
        {
            // graceful shutdown
        }
        finally
        {
            _consumer.Close();
        }
    }

    private async Task HandleMessageAsync(
        ConsumeResult<string, byte[]> result,
        CancellationToken ct)
    {
        EmailRequest request;
        try
        {
            request = MorentEnvelopeParser.ParseEmailSend(result.Message.Value);
        }
        catch (Exception ex)
        {
            _logger.LogWarning(
                ex,
                "Skipping invalid kafka message at offset {Offset}",
                result.Offset);
            _consumer!.Commit(result);
            return;
        }

        _logger.LogInformation(
            "Processing email from Kafka: correlationId={CorrelationId}, template={TemplateKey}, to={To}",
            request.CorrelationId,
            request.TemplateKey,
            request.To);

        using var scope = _scopeFactory.CreateScope();
        var renderer = scope.ServiceProvider.GetRequiredService<ITemplateRenderer>();
        var provider = scope.ServiceProvider.GetRequiredService<IEmailProvider>();
        var statusStore = scope.ServiceProvider.GetRequiredService<IEmailStatusStore>();

        try
        {
            if (await statusStore.IsProcessedAsync(request.CorrelationId, ct))
            {
                _logger.LogInformation(
                    "Email {CorrelationId} already processed, skipping",
                    request.CorrelationId);
                _consumer!.Commit(result);
                return;
            }

            await statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Rendering, ct: ct);
            var rendered = await renderer.RenderAsync(request, ct);
            await provider.SendAsync(rendered, ct);
            await statusStore.MarkStatusAsync(request.CorrelationId, EmailStatus.Sent, ct: ct);

            _logger.LogInformation(
                "Email {CorrelationId} sent successfully (queued from kafka)",
                request.CorrelationId);

            _consumer!.Commit(result);
        }
        catch (Exception ex) when (ex is not OperationCanceledException)
        {
            _logger.LogError(
                ex,
                "Failed to process kafka email {CorrelationId}",
                request.CorrelationId);

            await statusStore.MarkStatusAsync(
                request.CorrelationId,
                EmailStatus.Failed,
                ex.Message,
                ct);

            // не коммитим — повторная доставка
        }
    }

    public override void Dispose()
    {
        _consumer?.Dispose();
        base.Dispose();
    }
}
