using EmailService.Core.Models;

namespace EmailService.Core.Contracts;

public interface IEmailQueue
{
    ValueTask EnqueueAsync(EmailRequest request, CancellationToken ct = default);
    IAsyncEnumerable<EmailRequest> ReadAllAsync(CancellationToken ct = default);
}