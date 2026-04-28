namespace EmailService.Core.Models;

public record EmailRequest
{
    public string CorrelationId { get; init; } = Guid.NewGuid().ToString("N");
    public string To { get; init; } = null!;
    public string TemplateKey { get; init; } = null!;
    public IReadOnlyDictionary<string, string> Variables { get; init; } = new Dictionary<string, string>();
    public EmailPriority Priority { get; init; } = EmailPriority.Normal;
}

public enum EmailPriority { Low = 0, Normal = 1, High = 2 }