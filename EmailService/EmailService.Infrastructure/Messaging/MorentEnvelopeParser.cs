using System.Text.Json;
using EmailService.Core.Models;

namespace EmailService.Infrastructure.Messaging;

/// <summary>
/// Парсит envelope Morent и строит <see cref="EmailRequest"/>.
/// </summary>
public static class MorentEnvelopeParser
{
    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNameCaseInsensitive = false
    };

    /// <summary>
    /// Разбирает сырое Kafka-сообщение в запрос на отправку письма.
    /// </summary>
    public static EmailRequest ParseEmailSend(byte[] raw)
    {
        var envelope = JsonSerializer.Deserialize<MorentEnvelope>(raw, JsonOptions)
            ?? throw new InvalidOperationException("empty envelope");

        if (envelope.SchemaVersion != MorentEvents.SchemaVersion)
        {
            throw new InvalidOperationException(
                $"unsupported schema version: {envelope.SchemaVersion}");
        }

        if (envelope.Source != MorentEvents.SourceMorent)
        {
            throw new InvalidOperationException(
                $"unexpected event source: {envelope.Source}");
        }

        if (string.IsNullOrWhiteSpace(envelope.EventId)
            || string.IsNullOrWhiteSpace(envelope.EventType))
        {
            throw new InvalidOperationException("invalid envelope: missing event_id or event_type");
        }

        if (envelope.EventType != MorentEvents.EventEmailSend)
        {
            throw new InvalidOperationException(
                $"unsupported event type: {envelope.EventType}");
        }

        var payload = envelope.Data.Deserialize<EmailSendRequested>(JsonOptions)
            ?? throw new InvalidOperationException("empty email send payload");

        var to = payload.To.Trim().ToLowerInvariant();
        if (string.IsNullOrEmpty(to))
        {
            throw new InvalidOperationException("email recipient is empty");
        }

        var templateKey = payload.TemplateKey.Trim();
        if (string.IsNullOrEmpty(templateKey))
        {
            throw new InvalidOperationException("email template key is empty");
        }

        var correlationId = string.IsNullOrWhiteSpace(payload.CorrelationId)
            ? Guid.NewGuid().ToString("N")
            : payload.CorrelationId.Trim();

        return new EmailRequest
        {
            CorrelationId = correlationId,
            To = to,
            TemplateKey = templateKey,
            Variables = payload.Variables ?? new Dictionary<string, string>(),
            Priority = MapPriority(payload.Priority)
        };
    }

    private static EmailPriority MapPriority(string? priority) =>
        priority?.Trim().ToLowerInvariant() switch
        {
            "high" => EmailPriority.High,
            "low" => EmailPriority.Low,
            _ => EmailPriority.Normal
        };
}
