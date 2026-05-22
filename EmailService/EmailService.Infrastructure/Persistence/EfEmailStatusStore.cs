using EmailService.Core.Contracts;
using EmailService.Core.Models;
using EmailService.Infrastructure.Persistence.Entities;
using Microsoft.EntityFrameworkCore;

namespace EmailService.Infrastructure.Persistence;

/// <summary>
/// Entity Framework implementation of <see cref="IEmailStatusStore"/>.
/// </summary>
public class EfEmailStatusStore : IEmailStatusStore
{
    private readonly EmailDbContext _db;

    /// <summary>
    /// Initializes a new instance of the <see cref="EfEmailStatusStore"/> class.
    /// </summary>
    /// <param name="db">Database context.</param>
    public EfEmailStatusStore(EmailDbContext db)
    {
        _db = db;
    }

    /// <summary>
    /// Updates or creates an email status record.
    /// </summary>
    /// <param name="correlationId">Unique email correlation identifier.</param>
    /// <param name="status">Email processing status.</param>
    /// <param name="details">Optional status details.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    public async Task MarkStatusAsync(
        string correlationId,
        EmailStatus status,
        string? details = null,
        CancellationToken ct = default)
    {
        var log = await _db.EmailLogs.FindAsync([correlationId], ct);

        if (log != null)
        {
            log.Status = status;
            log.ErrorDetails = details;
            log.UpdatedAt = DateTime.UtcNow;
            log.Attempts++;

            _db.EmailLogs.Update(log);
        }
        else
        {
            log = new EmailLog
            {
                CorrelationId = correlationId,
                Status = status,
                ErrorDetails = details,
                CreatedAt = DateTime.UtcNow,
                UpdatedAt = DateTime.UtcNow,
                Attempts = 1
            };

            await _db.EmailLogs.AddAsync(log, ct);
        }

        await _db.SaveChangesAsync(ct);
    }

    /// <summary>
    /// Retrieves the current status of an email request.
    /// </summary>
    /// <param name="correlationId">Unique email correlation identifier.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns>Email status if found; otherwise <c>null</c>.</returns>
    public async Task<EmailStatus?> GetStatusAsync(
        string correlationId,
        CancellationToken ct = default)
    {
        var log = await _db.EmailLogs.FindAsync([correlationId], ct);

        return log?.Status;
    }

    /// <summary>
    /// Determines whether an email request has already been processed.
    /// </summary>
    /// <param name="correlationId">Unique email correlation identifier.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns>
    /// <c>true</c> if the email has already been processed; otherwise <c>false</c>.
    /// </returns>
    public async Task<bool> IsProcessedAsync(
        string correlationId,
        CancellationToken ct = default)
    {
        var status = await GetStatusAsync(correlationId, ct);

        return status is EmailStatus.Delivered
            or EmailStatus.Sent
            or EmailStatus.Spam
            or EmailStatus.Unsubscribed;
    }
}