namespace EmailService.Core.Models;

public record RenderedEmail
{
    public string CorrelationId { get; init; } = null!;
    public string To { get; init; } = null!;
    public string Subject { get; init; } = null!;
    public string HtmlContent { get; init; } = null!;
    public EmailPriority Priority { get; init; } = EmailPriority.Normal;
}