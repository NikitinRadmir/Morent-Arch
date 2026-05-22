using EmailService.Infrastructure.Messaging;

namespace EmailService.Infrastructure.Options;

/// <summary>
/// Настройки Kafka consumer для событий Morent.
/// </summary>
public class KafkaOptions
{
    /// <summary>
    /// Включить чтение из Kafka.
    /// </summary>
    public bool Enabled { get; set; }

    /// <summary>
    /// Список брокеров (через запятую).
    /// </summary>
    public string Brokers { get; set; } = "localhost:9092";

    /// <summary>
    /// Топик писем.
    /// </summary>
    public string Topic { get; set; } = MorentEvents.TopicEmails;

    /// <summary>
    /// Consumer group.
    /// </summary>
    public string GroupId { get; set; } = "email-service";
}
