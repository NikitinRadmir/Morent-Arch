namespace EmailService.Infrastructure.Options;

public class KafkaOptions
{
    public bool Enabled { get; set; }
    public string Brokers { get; set; } = "localhost:9092";
    public string Topic { get; set; } = "morent.emails";
    public string GroupId { get; set; } = "email-service";
}
