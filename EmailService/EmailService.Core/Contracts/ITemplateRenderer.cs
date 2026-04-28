using EmailService.Core.Models;

namespace EmailService.Core.Contracts;

public interface ITemplateRenderer
{
    Task<RenderedEmail> RenderAsync(EmailRequest request, CancellationToken ct = default);
}