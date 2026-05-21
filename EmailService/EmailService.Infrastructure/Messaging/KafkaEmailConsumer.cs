using System.Text.Json;
using Confluent.Kafka;
using EmailService.Core.Models;
using EmailService.Infrastructure.Options;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace EmailService.Infrastructure.Messaging;

public class KafkaEmailConsumer : BackgroundService
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = true
    };

    private readonly KafkaOptions _options;
    private readonly IServiceScopeFactory _scopeFactory;
    private readonly ILogger<KafkaEmailConsumer> _logger;
    private IConsumer<string, string>? _consumer;

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
        if (!_options.Enabled || string.IsNullOrWhiteSpace(_options.Brokers))
        {
            _logger.LogInformation("Kafka email consumer disabled");
            return;
        }

        var config = new ConsumerConfig
        {
            BootstrapServers = _options.Brokers,
            GroupId = _options.GroupId,
            AutoOffsetReset = AutoOffsetReset.Earliest,
            EnableAutoCommit = false
        };

        await KafkaTopicBootstrap.EnsureTopicAsync(
            _options.Brokers, _options.Topic, _logger, stoppingToken);

        _consumer = new ConsumerBuilder<string, string>(config).Build();
        _consumer.Subscribe(_options.Topic);
        _logger.LogInformation("Kafka email consumer started (topic {Topic}, group {Group})",
            _options.Topic, _options.GroupId);

        try
        {
            while (!stoppingToken.IsCancellationRequested)
            {
                try
                {
                    var result = _consumer.Consume(stoppingToken);
                    if (result?.Message?.Value == null)
                    {
                        continue;
                    }

                    if (await TryHandleAsync(result.Message.Value, stoppingToken))
                    {
                        _consumer.Commit(result);
                    }
                }
                catch (ConsumeException ex) when (ex.Error.Code == ErrorCode.UnknownTopicOrPart)
                {
                    _logger.LogWarning(
                        "Kafka topic {Topic} not ready yet, retrying in 3s (create a rental in Morent or wait for bootstrap)",
                        _options.Topic);
                    await Task.Delay(TimeSpan.FromSeconds(3), stoppingToken);
                }
                catch (ConsumeException ex)
                {
                    _logger.LogError(ex, "Kafka consume error");
                    await Task.Delay(TimeSpan.FromSeconds(1), stoppingToken);
                }
                catch (OperationCanceledException) when (stoppingToken.IsCancellationRequested)
                {
                    break;
                }
            }
        }
        finally
        {
            _consumer.Close();
            _logger.LogInformation("Kafka email consumer stopped");
        }
    }

    private async Task<bool> TryHandleAsync(string raw, CancellationToken ct)
    {
        MorentEnvelope? envelope;
        try
        {
            envelope = JsonSerializer.Deserialize<MorentEnvelope>(raw, JsonOptions);
        }
        catch (JsonException ex)
        {
            _logger.LogWarning(ex, "Invalid kafka message JSON, skipping");
            return true;
        }

        if (envelope == null ||
            envelope.SchemaVersion != MorentEventTypes.SchemaVersion ||
            envelope.Source != MorentEventTypes.SourceMorent ||
            string.IsNullOrWhiteSpace(envelope.EventType))
        {
            _logger.LogWarning("Unsupported kafka envelope, skipping");
            return true;
        }

        if (envelope.EventType != MorentEventTypes.EmailSend)
        {
            _logger.LogDebug("Ignoring event type {EventType}", envelope.EventType);
            return true;
        }

        MorentEmailSendData? data;
        try
        {
            data = envelope.Data.Deserialize<MorentEmailSendData>(JsonOptions);
        }
        catch (JsonException ex)
        {
            _logger.LogWarning(ex, "Invalid email.send payload, skipping");
            return true;
        }

        if (data == null ||
            string.IsNullOrWhiteSpace(data.To) ||
            string.IsNullOrWhiteSpace(data.TemplateKey))
        {
            _logger.LogWarning("Incomplete email.send payload, skipping");
            return true;
        }

        var correlationId = string.IsNullOrWhiteSpace(data.CorrelationId)
            ? Guid.NewGuid().ToString("N")
            : data.CorrelationId.Trim();

        var request = new EmailRequest
        {
            CorrelationId = correlationId,
            To = data.To.Trim(),
            TemplateKey = data.TemplateKey.Trim(),
            Variables = data.Variables ?? new Dictionary<string, string>(),
            Priority = ParsePriority(data.Priority)
        };

        try
        {
            using var scope = _scopeFactory.CreateScope();
            var ingestion = scope.ServiceProvider.GetRequiredService<MorentEmailIngestion>();
            await ingestion.IngestAsync(request, ct);
            return true;
        }
        catch (Exception ex) when (ex is not OperationCanceledException)
        {
            _logger.LogError(ex, "Failed to ingest kafka email {CorrelationId}", correlationId);
            return false;
        }
    }

    private static EmailPriority ParsePriority(string? priority)
    {
        if (string.IsNullOrWhiteSpace(priority))
        {
            return EmailPriority.Normal;
        }
        return Enum.TryParse<EmailPriority>(priority, ignoreCase: true, out var parsed)
            ? parsed
            : EmailPriority.Normal;
    }

    public override void Dispose()
    {
        _consumer?.Dispose();
        base.Dispose();
    }
}
