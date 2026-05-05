using EmailService.Core.Models;

namespace EmailService.Core.Contracts;

public interface IEmailStatusStore
{
    Task MarkStatusAsync(string correlationId, EmailStatus status, string? details = null, CancellationToken ct = default);
    Task<EmailStatus?> GetStatusAsync(string correlationId, CancellationToken ct = default);
    Task<bool> IsProcessedAsync(string correlationId, CancellationToken ct = default);
}