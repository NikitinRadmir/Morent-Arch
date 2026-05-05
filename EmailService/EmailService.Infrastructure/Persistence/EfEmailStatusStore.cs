using EmailService.Core.Contracts;
using EmailService.Core.Models;
using EmailService.Infrastructure.Persistence.Entities;
using Microsoft.EntityFrameworkCore;

namespace EmailService.Infrastructure.Persistence;

public class EfEmailStatusStore : IEmailStatusStore
{
    private readonly EmailDbContext _db;

    public EfEmailStatusStore(EmailDbContext db) => _db = db;

    public async Task MarkStatusAsync(string correlationId, EmailStatus status, string? details = null, CancellationToken ct = default)
    {
        var log = await _db.EmailLogs.FindAsync([correlationId], ct)
                  ?? new EmailLog { CorrelationId = correlationId };

        log.Status = status;
        log.ErrorDetails = details;
        log.UpdatedAt = DateTime.UtcNow;
        if (log.Attempts == 0) log.Attempts = 1;

        _db.EmailLogs.Update(log);
        await _db.SaveChangesAsync(ct);
    }

    public async Task<EmailStatus?> GetStatusAsync(string correlationId, CancellationToken ct = default)
    {
        var log = await _db.EmailLogs.FindAsync([correlationId], ct);
        return log?.Status;
    }

    public async Task<bool> IsProcessedAsync(string correlationId, CancellationToken ct = default)
    {
        var status = await GetStatusAsync(correlationId, ct);
        return status is EmailStatus.Delivered or EmailStatus.Sent or EmailStatus.Spam or EmailStatus.Unsubscribed;
    }
}