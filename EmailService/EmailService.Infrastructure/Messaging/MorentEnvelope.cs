using System.Text.Json;
using System.Text.Json.Serialization;

namespace EmailService.Infrastructure.Messaging;

public sealed class MorentEnvelope
{
    [JsonPropertyName("event_id")]
    public string EventId { get; set; } = "";

    [JsonPropertyName("event_type")]
    public string EventType { get; set; } = "";

    [JsonPropertyName("schema_version")]
    public string SchemaVersion { get; set; } = "";

    [JsonPropertyName("timestamp")]
    public DateTime Timestamp { get; set; }

    [JsonPropertyName("source")]
    public string Source { get; set; } = "";

    [JsonPropertyName("data")]
    public JsonElement Data { get; set; }
}

public sealed class MorentEmailSendData
{
    [JsonPropertyName("correlationId")]
    public string CorrelationId { get; set; } = "";

    [JsonPropertyName("to")]
    public string To { get; set; } = "";

    [JsonPropertyName("templateKey")]
    public string TemplateKey { get; set; } = "";

    [JsonPropertyName("variables")]
    public Dictionary<string, string>? Variables { get; set; }

    [JsonPropertyName("priority")]
    public string? Priority { get; set; }
}

public static class MorentEventTypes
{
    public const string SchemaVersion = "1";
    public const string SourceMorent = "morent-backend";
    public const string EmailSend = "morent.email.send";
}
