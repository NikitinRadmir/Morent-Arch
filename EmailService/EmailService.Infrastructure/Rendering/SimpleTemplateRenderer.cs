using System.Net;
using System.Text.RegularExpressions;
using EmailService.Core.Contracts;
using EmailService.Core.Exceptions;
using EmailService.Core.Models;
using EmailService.Infrastructure.Options;
using Microsoft.Extensions.Caching.Memory;
using Microsoft.Extensions.Options;

namespace EmailService.Infrastructure.Rendering;

/// <summary>
/// Renders email templates using simple HTML-encoded placeholders.
/// </summary>
public partial class SimpleTemplateRenderer : ITemplateRenderer
{
    private readonly TemplatesOptions _options;
    private readonly IMemoryCache _cache;

    private static readonly Regex SubjectRegex = MySubjectRegex();

    /// <summary>
    /// Initializes a new instance of the <see cref="SimpleTemplateRenderer"/> class.
    /// </summary>
    /// <param name="options">Template rendering configuration options.</param>
    /// <param name="cache">Memory cache instance.</param>
    public SimpleTemplateRenderer(
        IOptions<TemplatesOptions> options,
        IMemoryCache cache)
    {
        _options = options.Value;
        _cache = cache;
    }

    /// <summary>
    /// Renders an email template using the provided request data.
    /// </summary>
    /// <param name="request">Email request containing template information.</param>
    /// <param name="ct">Cancellation token for the operation.</param>
    /// <returns>A rendered email instance.</returns>
    /// <exception cref="TemplateNotFoundException">
    /// Thrown when the template file cannot be found.
    /// </exception>
    public async Task<RenderedEmail> RenderAsync(
        EmailRequest request,
        CancellationToken ct = default)
    {
        var templateKey = request.TemplateKey;

        var filePath = Path.Combine(
            _options.BasePath,
            templateKey + _options.DefaultExtension);

        if (!File.Exists(filePath))
        {
            throw new TemplateNotFoundException(templateKey);
        }

        var cacheKey = $"tpl:{templateKey}";

        var templateHtml = await _cache.GetOrCreateAsync(cacheKey, async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow =
                TimeSpan.FromSeconds(_options.CacheDurationSeconds);

            return await File.ReadAllTextAsync(filePath, ct);
        }) ?? throw new InvalidOperationException($"Template '{templateKey}' could not be loaded.");

        var renderedHtml = RenderVariables(templateHtml, request.Variables);
        var subjectMatch = SubjectRegex.Match(renderedHtml);

        var subject = subjectMatch.Success
            ? subjectMatch.Groups[1].Value.Trim()
            : "Без темы";

        var cleanHtml = SubjectRegex
            .Replace(renderedHtml, "")
            .Trim();

        return new RenderedEmail
        {
            CorrelationId = request.CorrelationId,
            To = request.To,
            Subject = subject,
            HtmlContent = cleanHtml,
            Priority = request.Priority
        };
    }

    /// <summary>
    /// Creates a regular expression for extracting the email subject.
    /// </summary>
    /// <returns>Compiled regular expression instance.</returns>
    [GeneratedRegex(@"<subject>(.*?)</subject>", RegexOptions.Singleline)]
    private static partial Regex MySubjectRegex();

    private static string RenderVariables(
        string template,
        IReadOnlyDictionary<string, string> variables)
    {
        return VariableRegex().Replace(template, match =>
        {
            var key = match.Groups[1].Value.Trim();

            return variables.TryGetValue(key, out var value)
                ? WebUtility.HtmlEncode(value)
                : string.Empty;
        });
    }

    [GeneratedRegex(@"\{\{\s*([A-Za-z0-9_.-]+)\s*\}\}", RegexOptions.CultureInvariant)]
    private static partial Regex VariableRegex();
}
