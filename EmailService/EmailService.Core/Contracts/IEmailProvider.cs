using EmailService.Core.Models;

namespace EmailService.Core.Contracts;

public interface IEmailProvider
{
    Task SendAsync(RenderedEmail email, CancellationToken ct = default);
}