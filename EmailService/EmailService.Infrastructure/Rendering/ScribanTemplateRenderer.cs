using System.Text.RegularExpressions;
using EmailService.Core.Contracts;
using EmailService.Core.Exceptions;
using EmailService.Core.Models;
using EmailService.Infrastructure.Options;
using Microsoft.Extensions.Caching.Memory;
using Microsoft.Extensions.Options;
using Scriban;

namespace EmailService.Infrastructure.Rendering;

public partial class ScribanTemplateRenderer : ITemplateRenderer
{
    private readonly TemplatesOptions _options;
    private readonly IMemoryCache _cache;
    private static readonly Regex SubjectRegex = MySubjectRegex();

    public ScribanTemplateRenderer(IOptions<TemplatesOptions> options, IMemoryCache cache)
    {
        _options = options.Value;
        _cache = cache;
    }

    public async Task<RenderedEmail> RenderAsync(EmailRequest request, CancellationToken ct = default)
    {
        var templateKey = request.TemplateKey;
        var filePath = Path.Combine(_options.BasePath, templateKey + _options.DefaultExtension);

        if (!File.Exists(filePath))
            throw new TemplateNotFoundException(templateKey);

        var cacheKey = $"tpl:{templateKey}";
        var template = await _cache.GetOrCreateAsync(cacheKey, async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromSeconds(_options.CacheDurationSeconds);
            var html = await File.ReadAllTextAsync(filePath, ct);
            return Template.Parse(html);
        })!;

        var renderedHtml = await template.RenderAsync(request.Variables, memberRenamer: member => member.Name);
        var subjectMatch = SubjectRegex.Match(renderedHtml);
        var subject = subjectMatch.Success ? subjectMatch.Groups[1].Value.Trim() : "Без темы";
        var cleanHtml = SubjectRegex.Replace(renderedHtml, "").Trim();

        return new RenderedEmail
        {
            CorrelationId = request.CorrelationId,
            To = request.To,
            Subject = subject,
            HtmlContent = cleanHtml,
            Priority = request.Priority
        };
    }

    [GeneratedRegex(@"<subject>(.*?)</subject>", RegexOptions.Singleline)]
    private static partial Regex MySubjectRegex();
}