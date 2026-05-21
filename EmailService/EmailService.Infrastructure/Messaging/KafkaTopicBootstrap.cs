using Confluent.Kafka;
using Confluent.Kafka.Admin;
using Microsoft.Extensions.Logging;

namespace EmailService.Infrastructure.Messaging;

internal static class KafkaTopicBootstrap
{
    public static async Task EnsureTopicAsync(string brokers, string topic, ILogger logger, CancellationToken ct = default)
    {
        if (string.IsNullOrWhiteSpace(brokers) || string.IsNullOrWhiteSpace(topic))
        {
            return;
        }

        using var admin = new AdminClientBuilder(new AdminClientConfig
        {
            BootstrapServers = brokers
        }).Build();

        try
        {
            var metadata = admin.GetMetadata(topic, TimeSpan.FromSeconds(10));
            var exists = metadata.Topics.Any(t =>
                t.Topic == topic && t.Error.Code == ErrorCode.NoError);

            if (exists)
            {
                return;
            }
        }
        catch (KafkaException ex)
        {
            logger.LogDebug(ex, "Could not read metadata for topic {Topic}, will try to create", topic);
        }

        try
        {
            await admin.CreateTopicsAsync(
            [
                new TopicSpecification
                {
                    Name = topic,
                    NumPartitions = 1,
                    ReplicationFactor = 1
                }
            ]);
            logger.LogInformation("Kafka topic {Topic} created", topic);
        }
        catch (CreateTopicsException ex) when (
            ex.Results.Count > 0 &&
            ex.Results[0].Error.Code == ErrorCode.TopicAlreadyExists)
        {
            // race with another service
        }
    }
}
