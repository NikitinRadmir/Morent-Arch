using EmailService.Core.Models;

namespace EmailService.Infrastructure.Persistence.Entities;

public class EmailLog
{
    public string CorrelationId { get; set; } = null!;
    public EmailStatus Status { get; set; }
    public string? ProviderMessageId { get; set; }
    public int Attempts { get; set; }
    public string? ErrorDetails { get; set; }
    public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
    public DateTime? UpdatedAt { get; set; }
}