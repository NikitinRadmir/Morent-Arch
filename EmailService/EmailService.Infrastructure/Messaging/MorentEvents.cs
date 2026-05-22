using System.Text.Json;
using System.Text.Json.Serialization;

namespace EmailService.Infrastructure.Messaging;

/// <summary>
/// Константы событий Morent (shared/morent-events).
/// </summary>
public static class MorentEvents
{
    public const string SchemaVersion = "1";
    public const string SourceMorent = "morent-backend";
    public const string TopicEmails = "morent.emails";
    public const string EventEmailSend = "morent.email.send";
    public const string TemplateWelcomeRegistered = "welcome_registered";
}

/// <summary>
/// Обёртка Kafka-события Morent.
/// </summary>
public sealed class MorentEnvelope
{
    [JsonPropertyName("event_id")]
    public string EventId { get; init; } = "";

    [JsonPropertyName("event_type")]
    public string EventType { get; init; } = "";

    [JsonPropertyName("schema_version")]
    public string SchemaVersion { get; init; } = "";

    [JsonPropertyName("timestamp")]
    public DateTime Timestamp { get; init; }

    [JsonPropertyName("source")]
    public string Source { get; init; } = "";

    [JsonPropertyName("data")]
    public JsonElement Data { get; init; }
}

/// <summary>
/// Payload <c>morent.email.send</c>.
/// </summary>
public sealed class EmailSendRequested
{
    [JsonPropertyName("correlationId")]
    public string CorrelationId { get; init; } = "";

    [JsonPropertyName("to")]
    public string To { get; init; } = "";

    [JsonPropertyName("templateKey")]
    public string TemplateKey { get; init; } = "";

    [JsonPropertyName("variables")]
    public Dictionary<string, string>? Variables { get; init; }

    [JsonPropertyName("priority")]
    public string? Priority { get; init; }
}
